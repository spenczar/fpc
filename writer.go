package fpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	DefaultCompression = 10
	MaxCompression     = 0xFF

	floatChunkSize = 8
)

// A Writer is an io.WriteCloser which FPC-compresses data it receives
// and writes it to an underlying writer, w.  Writes to a Writer are
type Writer struct {
	w     io.Writer
	level int
	enc   *blockEncoder

	wroteHeader bool
	closed      bool
}

// NewWriter makes a new Writer which writes compressed data to w
// using the default compression level.
func NewWriter(w io.Writer) *Writer {
	z, _ := NewWriterLevel(w, DefaultCompression)
	return z
}

// NewWriterLevel makes a new Writer which writes compressed data to w
// using a provided compression level. Higher compression levels will
// result in more compressed data, but require exponentially more
// memory. The space required is O(2^level) bytes. NewWriterLevel
// returns an error if an invalid compression level is provided.
func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	if level < 1 || level > MaxCompression {
		return nil, fmt.Errorf("fpc: invalid compression level: %d", level)
	}
	z := &Writer{
		w:     w,
		level: level,
		enc:   newBlockEncoder(w, uint(level)),
	}
	return z, nil
}

// Write interprets b as a stream of byte-encoded, 64-bit IEEE 754
// floating point values. The length of b must be a multiple of 8 in
// order to match this expectation.
func (w *Writer) Write(b []byte) (int, error) {
	if len(b)%8 != 0 {
		return 0, errors.New("fpc.Write: len of data must be a multiple of 8")
	}
	for i := 0; i < len(b); i += 8 {
		if err := w.writeBytes(b[i : i+8]); err != nil {
			return i, err
		}
	}
	return len(b), nil
}

// WriteFloat writes a single float64 value to the encoded stream.
func (w *Writer) WriteFloat(f float64) error {
	return w.writeFloat64(f)
}

// Flush will make sure all internally-buffered values are written to
// w. FPC's format specifies that data get written in blocks; calling
// Flush will write the current data to a block, even if it results in
// a partial block.
//
// Flush does not flush the underlying io.Writer which w is delegating
// to.
func (w *Writer) Flush() error {
	if err := w.ensureHeader(); err != nil {
		return err
	}
	return w.enc.flush()
}

// Close will flush the Writer and make any subsequent writes return
// errors. It does not close the underlying io.Writer which w is
// delegating to.
func (w *Writer) Close() error {
	if w.closed == true {
		return nil
	}
	w.closed = true
	return w.Flush()
}

func (w *Writer) ensureHeader() error {
	if !w.wroteHeader {
		w.wroteHeader = true
		_, err := w.w.Write([]byte{byte(w.level)})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeFloat64(f float64) error {
	return w.writeUint64(math.Float64bits(f))
}

func (w *Writer) writeUint64(u uint64) error {
	if err := w.ensureHeader(); err != nil {
		return err
	}
	if err := w.enc.encode(u); err != nil {
		return err
	}
	return nil
}

// writeBytes writes a single 8-byte encoded IEEE 754 float
func (w *Writer) writeBytes(b []byte) error {
	if err := w.writeUint64(binary.LittleEndian.Uint64(b)); err != nil {
		return err
	}
	return nil
}
