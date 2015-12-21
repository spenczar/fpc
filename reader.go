package fpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// A DataError is returned when the FPC data is found to be syntactically
// invalid.
type DataError string

func (e DataError) Error() string {
	return "fpc data invalid: " + string(e)
}

// A Reader provides io.Reader-style access to a stream of FPC
// compressed data.
type Reader struct {
	r io.Reader

	fcm  predictor
	dfcm predictor

	initialized bool
	eof         bool

	block block // Current block being read
}

// NewReader creates a new Reader which reads and decompresses FPC data from
// the given io.Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

func (r *Reader) initialize() (err error) {
	comp, err := r.readGlobalHeader()
	if err != nil {
		return err
	}
	tableSize := uint(1 << comp)
	r.fcm = newFCM(tableSize)
	r.dfcm = newDFCM(tableSize)
	r.initialized = true
	return nil
}

// readGlobalHeader reads one byte and parses it as the compression level.
func (r *Reader) readGlobalHeader() (comp uint, err error) {
	var b []byte = make([]byte, 1)
	n, err := r.r.Read(b)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, DataError("missing first byte compression header")
	}
	return uint(b[0]), nil
}

// Read reads from up to (len(buf) / 8) IEEE 754 64-bit floating point
// values into buf. It is an error to provide a buf whose length is
// not a multiple of 8, because that would prevent encoding of the
// read float64s.
//
// If more values might be available, Read will return len(buf),
// nil. If no more values are available, Read will return with
// err==io.EOF
func (r *Reader) Read(buf []byte) (int, error) {
	if len(buf)%8 != 0 {
		return 0, errors.New("fpc: []byte passed to Reader.Read must have length which is a multiple of 8")
	}

	if !r.initialized {
		err := r.initialize()
		if err != nil {
			return 0, err
		}
	}

	nRead := 0
	for {
		// If available, read data from the block.
		n, err := r.readFromBlock(buf)
		if err != nil {
			return n, err
		}
		nRead += n
		// We've read everything we need to.
		if nRead == len(buf) {
			return nRead, nil
		}

		// End of block.
		if n == 0 {
			// Check whether counts match up.
			if r.block.nRecRead != r.block.nRec {
				return nRead, DataError("block record length too short")
			}
			if r.block.nByteRead != r.block.nByte {
				return nRead, DataError(fmt.Sprintf("block byte length too short, have=%d  want=%d", r.block.nByteRead, r.block.nByte))
			}

			// Find a new block
			r.block, err = r.readBlockHeader()
			if err != nil {
				return nRead, err
			}
		}
	}
}

// ReadFloats will read data from the underlying io.Reader, parsing
// the data it gets back as float64s and putting them into fs. If no
// more values are available, ReadFloats will returns with an
// err==io.EOF.
func (r *Reader) ReadFloats(fs []float64) (int, error) {
	buf := make([]byte, 8)
	var val uint64
	for i := range fs {
		_, err := r.Read(buf)
		if err != nil {
			return i, err
		}
		val = binary.LittleEndian.Uint64(buf)
		fs[i] = math.Float64frombits(val)
	}
	return len(fs), nil
}

// ReadFloat will read data from the underlying io.Reader until it has
// read enough data to provide a float64, decodes that data, and
// returns the decoded float64. If an error is encountered while
// reading, it returns 0 and that error. If no more values are
// available, ReadFloat will return with err==io.EOF.
func (r *Reader) ReadFloat() (float64, error) {
	buf := make([]byte, 8)
	_, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	val := binary.LittleEndian.Uint64(buf)
	return math.Float64frombits(val), nil
}

// readBlockHeader reads the block header and record headers that start a data
// block. It returns the slice of record headers, the number of bytes remaining
// in the block, and any errors encountered while reading.
func (r *Reader) readBlockHeader() (b block, err error) {
	// The first 6 bytes of the block describe the number of records and bytes
	// in the block.
	buf := make([]byte, 6)
	n, err := r.r.Read(buf)
	if n == 0 && err == io.EOF {
		// No data available: This is a genuine EOF. We have no blocks left.
		return b, io.EOF
	} else if n < len(buf) || err == io.EOF {
		// Partial data available: This is a corrupted header, we expected 6 bytes.
		return b, DataError("block header too short")
	} else if err != nil {
		// Some other unexpected error
		return b, err
	}
	b.nRec, b.nByte = decodeBlockHeader(buf)
	b.nByteRead += 6 // the first 6 bytes are included in the header's count

	// Each record has a 4-bit header value. These headers have 1 bit to
	// describe which predictor hash table to use, and 3 bits to describe how
	// many zero bits prefix their associated value.
	//
	// The 4-bit records are packed as pairs into bytes. If there are an odd
	// number of records in the block, then the last 4-bit header is
	// meaningless and can be discarded.
	b.headers = make([]header, b.nRec)

	// Read out the appropriate number of bytes.
	buf = make([]byte, b.nRec/2)
	n, err = io.ReadFull(r.r, buf)
	if err != nil {
		return b, err
	}
	for i, byte := range buf {
		b.headers[2*i], b.headers[2*i+1] = decodeHeaders(byte)
	}
	b.nByteRead += b.nRec / 2

	// If there are an odd number of records, then read just the first 4 bits
	// of the next byte.
	if b.nRec%2 == 1 {
		// Read one byte.
		buf = buf[:1]
		_, err = io.ReadFull(r.r, buf)
		if err != nil {
			return b, err
		}
		b.headers[b.nRec-1], _ = decodeHeaders(buf[0])
		b.nByteRead += 1
	}

	return b, nil
}

func (r *Reader) readFromBlock(p []byte) (int, error) {
	var (
		b    []byte // workspace for decoding
		val  uint64
		pred uint64
		h    header

		bytesDecoded int
	)

	b = make([]byte, 8) // records can be at most 8 bytes
	for r.block.nRecRead < r.block.nRec && len(p) > 0 {
		// Get as many bytes off the reader as the header says we should take.
		h = r.block.headers[r.block.nRecRead]
		n, err := r.r.Read(b[:h.len])
		if n < int(h.len) || err == io.EOF {
			return bytesDecoded, DataError("missing records")
		}
		if err != nil {
			return bytesDecoded, err
		}

		// Parse the bytes.
		val = decodeData(b[:h.len])

		// XOR with the predictions to get back the true values.
		if h.pType == fcmPredictor {
			pred = r.fcm.predict()
		} else {
			pred = r.dfcm.predict()
		}
		val = pred ^ val
		r.fcm.update(val)
		r.dfcm.update(val)

		// Write the value to p.
		binary.LittleEndian.PutUint64(p[:8], val)
		p = p[8:]

		// increment counters
		bytesDecoded += 8
		r.block.nByteRead += int(h.len)
		r.block.nRecRead += 1
	}
	return bytesDecoded, nil
}

type block struct {
	headers []header

	// Counters for current position within the block
	nRecRead  int
	nByteRead int

	// Total counts for the block
	nRec  int
	nByte int
}
