package fpc

import (
	"encoding/binary"
	"io"
	"math"
)

const (
	encoderBuffer      = 1024 // initial size for encoder buffer
	defaultCompression = 10
	defaultBlockSize   = 32768

	blockHeaderSize = 6 // in bytes
)

var byteOrder = binary.LittleEndian

type blockEncoder struct {
	buf       []byte // Internal buffer of encoded values
	blockSize int    // size of blocks in bytes

	w   io.Writer // Destination for encoded bytes
	enc *encoder  // Underlying machinery for encoding pairs of floats

	// Mutable state below
	last     float64 // last value received to encode
	nRecords int     // Count of float64s received in this block
}

func newBlockEncoder(w io.Writer, compression uint) *blockEncoder {
	return &blockEncoder{
		buf:       make([]byte, 0, defaultBlockSize),
		blockSize: defaultBlockSize,
		w:         w,
		enc:       newEncoder(compression),
		last:      0,
		nRecords:  0,
	}
}

func (b *blockEncoder) encode(f float64) error {
	// Encode floats in pairs
	if b.nRecords%2 == 1 {
		b.last = f
		b.nRecords += 1
		return nil
	}

	encoded := b.enc.encode(math.Float64bits(b.last), math.Float64bits(f))

	// If the encoded data would overflow our buffer, then flush first
	if len(encoded)+len(b.buf) > cap(b.buf) {
		if err := b.flush(); err != nil {
			return err
		}
	}

	// Append data to the block
	b.buf = append(b.buf, encoded...)
	b.nRecords += 1

	return nil
}

func (b *blockEncoder) flush() error {
	// Prepend data with header
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
	b.buf = make([]byte, 0, b.blockSize)
	b.nRecords = 0
	return nil
}

func (b *blockEncoder) encodeBlock() []byte {
	// The header is layed out as two little-endian 24-bit unsigned
	// integers. The first integer is the number of records in the block, and
	// the second is the number of bytes.
	h := make([]byte, blockHeaderSize)

	//First three bytes are the number of records in the block.
	h[0] = byte(b.nRecords)
	h[1] = byte(b.nRecords >> 8)
	h[2] = byte(b.nRecords >> 16)

	// Next three bytes are the number of bytes in the block.
	nByte := len(b.buf) + blockHeaderSize
	h[3] = byte(nByte)
	h[4] = byte(nByte >> 8)
	h[5] = byte(nByte >> 16)

	// After the header is all the rest of the data.
	return append(h, b.buf...)
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

// compute the difference between v and the best predicted value; return that
// difference and which predictor was the most effective. Updates predictors as
// a side effect.
func (e *encoder) computeDiff(v uint64) (d uint64, p predictorClass) {
	fcmDelta := e.fcm.predict() ^ v
	e.fcm.update(v)

	dfcmDelta := e.dfcm.predict() ^ v
	e.dfcm.update(v)

	if fcmDelta <= dfcmDelta {
		return fcmDelta, fcmPredictor
	} else {
		return dfcmDelta, dfcmPredictor
	}
}

// encode a pair of values
func (e *encoder) encode(v1, v2 uint64) []byte {
	d1, p1 := e.computeDiff(v1)
	d2, p2 := e.computeDiff(v2)

	d1Prefix, d1NZero := e.prefixCode(d1, p1)
	d2Prefix, d2NZero := e.prefixCode(d2, p2)

	// First byte contains prefixes for the values
	e.buf[0] = d1Prefix<<4 | d2Prefix

	e.encodeNonzero(d1, d1NZero, e.buf[1:8-d1NZero+1])
	e.encodeNonzero(d2, d2NZero, e.buf[1+(8-d1NZero):])
	return e.buf[:1+(8-d1NZero)+(8-d2NZero)]
}

func (e *encoder) encodeNonzero(v uint64, nZero uint64, into []byte) {
	// Starting with the first nonzero byte, copy v's data into the byte slice.
	//
	// Unrolling this loop into a switch speeds up the computation dramatically.

	switch nZero {
	case 0:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
		into[6] = byte((v >> 48) & 0xFF)
		into[7] = byte((v >> 56) & 0xFF)
	case 1:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
		into[6] = byte((v >> 48) & 0xFF)
	case 2:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
		into[5] = byte((v >> 40) & 0xFF)
	case 3:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 32) & 0xFF)
	case 4:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
	case 5:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
	case 6:
		into[0] = byte(v & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
	case 7:
		into[0] = byte(v & 0xFF)
	}
}

// Compute 4-bit header for the value. The first bit tells which predictor was
// used; the next three bits tell how many leading zero bytes there are for the
// value.
func (e *encoder) prefixCode(v uint64, p predictorClass) (code uint8, nZeroBytes uint64) {
	z := clzBytes(v)

	// We want to encode the number of zero bytes into 3 bits. 4-zero-byte
	// prefixes are extremely rare, so they are treated like 3-zero-byte
	// prefixes. Burtscher and Ratanaworabhan explain in "FPC: A High-Speed
	// Compressor for Double-Precision Floating-Point Data:"
	//
	//   "Since there can be between zero and eight leading zero bytes, i.e.,
	//   nine possibilities, not all of them can be encoded with a three-bit
	//   value. We decided not to support a leading zero count of four because
	//   it occurs only rarely (cf. Section 5.4). Consequently, all xor results
	//   with four leading zero bytes are treated like values with only three
	//   leading zero bytes and the fourth zero byte is emitted as part of the
	//   residual."
	zOrig := z
	if z > 4 {
		z -= 1
	} else if z == 4 {
		z -= 1
		zOrig -= 1
	}
	if p == fcmPredictor {
		return uint8(z), zOrig
	} else {
		return uint8(z | 0x8), zOrig
	}
}

func (e *encoder) choosePredictor(v uint64) predictorClass {
	return fcmPredictor
}
