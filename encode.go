package fpc

import (
	"encoding/binary"
	"io"
	"math"
)

const (
	maxRecordsPerBlock = 32768
	blockHeaderSize    = 6 // in bytes
)

var byteOrder = binary.LittleEndian

// pairHeader combines the headers for two values into a single byte
type pairHeader struct {
	h1 header
	h2 header
}

func (ph pairHeader) encode() byte {
	return (ph.h1.encode()<<4 | ph.h2.encode())
}

// header is a container for the count of the number of non-zero bytes in an
// encoded value, and the type of predictor used to generate the encoded value
type header struct {
	len   uint8
	pType predictorClass
}

// the top bit is the predictor type bit. Bottom 3 bits encode the number of
// leading zero bytes for the value.
func (h header) encode() byte {
	if h.len > 4 {
		return byte(h.pType)<<3 | byte(h.len-1)
	} else {
		return byte(h.pType)<<3 | byte(h.len)
	}
}

type blockEncoder struct {
	headers          []byte
	values           []byte
	buffer           []byte
	compressionLevel uint

	w   io.Writer // Destination for encoded bytes
	enc *encoder  // Underlying machinery for encoding pairs of floats

	// Mutable state below
	last     uint64 // last value received to encode
	nRecords int    // Count of float64s received in this block
	nBytes   int    // Count of bytes in this block
}

func newBlockEncoder(w io.Writer, compression uint) *blockEncoder {
	return &blockEncoder{
		headers:          make([]byte, 0, maxRecordsPerBlock),
		values:           make([]byte, 0, maxRecordsPerBlock*8),
		compressionLevel: compression,
		w:                w,
		enc:              newEncoder(compression),
		last:             0,
		nRecords:         0,
	}
}

// reset the encoder to use a new io.Writer and compression level.
// This is faster and more efficient than creating a new encoder.
func (b *blockEncoder) reset(w io.Writer, compression uint) error {
	// Flush anything hanging around in the buffer
	err := b.flush()
	if err != nil {
		return err
	}

	// The actual reset
	b.headers = b.headers[:0]
	b.values = b.values[:0]

	// Only create a new encoder if we have to, otherwise re-use what we already have
	if compression == b.compressionLevel {
		b.enc.clear()
	} else {
		b.enc = newEncoder(compression)
		b.compressionLevel = compression
	}

	b.w = w
	b.last = 0
	b.nRecords = 0
	b.nBytes = 0

	return nil
}

func (b *blockEncoder) encode(v uint64) error {
	// Encode values in pairs
	if b.nRecords%2 == 0 {
		b.last = v
		b.nRecords += 1
		return nil
	}
	header, data := b.enc.encode(b.last, v)
	nBytes := 1 + len(data) // 1 for header

	// Append data to the block
	b.headers = append(b.headers, header.encode())
	b.values = append(b.values, data...)
	b.nRecords += 1
	b.nBytes += nBytes

	// Flush if we need to
	if b.nRecords == maxRecordsPerBlock {
		if err := b.flush(); err != nil {
			return err
		}
	}
	return nil

}

func (b *blockEncoder) encodeFloat(f float64) error {
	return b.encode(math.Float64bits(f))
}

func (b *blockEncoder) flush() error {
	if b.nRecords == 0 {
		return nil
	}
	if b.nRecords%2 == 1 {
		// There's an extra record waiting for a partner. Add a dummy value by
		// encoding a zero and adding it to data.
		h, data := b.enc.encode(b.last, 0)
		// Truncate out the dummy value's data. The header remains, but it
		// won't do any harm.
		data = data[:h.h1.len]
		b.headers = append(b.headers, h.encode())
		b.values = append(b.values, data...)
	}

	block := b.encodeBlock()
	// Write data out
	n, err := b.w.Write(block)
	if err != nil {
		return err
	}
	if n < len(block) {
		return io.ErrShortWrite
	}

	// Reset buffer and counters
	b.headers = b.headers[:0]
	b.values = b.values[:0]
	b.nRecords = 0
	b.nBytes = 0
	return nil
}

func (b *blockEncoder) encodeBlock() []byte {
	// The block header is laid out as two little-endian 24-bit unsigned
	// integers. The first integer is the number of records in the block, and
	// the second is the number of bytes.
	nByte := len(b.headers) + len(b.values) + blockHeaderSize

	if b.buffer == nil {
		b.buffer = make([]byte, 6, nByte)
	}
	block := b.buffer[:6]

	//First three bytes are the number of records in the block.
	block[0] = byte(b.nRecords)
	block[1] = byte(b.nRecords >> 8)
	block[2] = byte(b.nRecords >> 16)

	// Next three bytes are the number of bytes in the block.
	block[3] = byte(nByte)
	block[4] = byte(nByte >> 8)
	block[5] = byte(nByte >> 16)

	// Record headers follow the block header
	block = append(block, b.headers...)

	// After the header is all the rest of the data.
	block = append(block, b.values...)
	b.buffer = block

	return block
}

type encoder struct {
	buf []byte

	// predictors
	fcm  predictor
	dfcm predictor
}

func newEncoder(compression uint) *encoder {
	tableSize := uint(1 << compression)
	return &encoder{
		buf:  make([]byte, 17),
		fcm:  newFCM(tableSize),
		dfcm: newDFCM(tableSize),
	}
}

func (e *encoder) clear() {
	clear(e.buf)
	e.fcm.clear()
	e.dfcm.clear()
}

// compute the difference between v and the best predicted value; return that
// difference and which predictor was the most effective. Updates predictors as
// a side effect.
func (e *encoder) computeDiff(v uint64) (d uint64, h header) {
	fcmDelta := e.fcm.predict() ^ v
	e.fcm.update(v)

	dfcmDelta := e.dfcm.predict() ^ v
	e.dfcm.update(v)

	if fcmDelta <= dfcmDelta {
		d = fcmDelta
		h.pType = fcmPredictor
	} else {
		d = dfcmDelta
		h.pType = dfcmPredictor
	}
	h.len = uint8(8 - clzBytes(d))

	//   "Since there can be between zero and eight leading zero bytes, i.e.,
	//   nine possibilities, not all of them can be encoded with a three-bit
	//   value. We decided not to support a leading zero count of four because
	//   it occurs only rarely (cf. Section 5.4). Consequently, all xor results
	//   with four leading zero bytes are treated like values with only three
	//   leading zero bytes and the fourth zero byte is emitted as part of the
	//   residual."
	//
	// Here we add 1, to include one of the leading 0s in the residual.
	if h.len == 4 {
		h.len += 1
	}
	return d, h
}

// encode a pair of values
func (e *encoder) encode(v1, v2 uint64) (h pairHeader, data []byte) {
	d1, h1 := e.computeDiff(v1)
	d2, h2 := e.computeDiff(v2)

	h = pairHeader{h1, h2}

	e.encodeNonzero(d1, h1.len, e.buf[:h1.len])
	e.encodeNonzero(d2, h2.len, e.buf[h1.len:h1.len+h2.len])
	return h, e.buf[:h1.len+h2.len]
}

func (e *encoder) encodeNonzero(v uint64, n uint8, into []byte) {
	// Starting with the first nonzero byte, copy v's data into the byte slice.
	//
	// Unrolling this loop into a switch speeds up the computation dramatically.
	switch n {
	case 8:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
		into[6] = byte((v >> 48) & 0xFF)
		into[7] = byte((v >> 56) & 0xFF)
	case 7:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
		into[6] = byte((v >> 48) & 0xFF)
	case 6:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
	case 5:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
	case 3:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
	case 2:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
	case 1:
		into[0] = byte(v & 0xFF)
	}
}
