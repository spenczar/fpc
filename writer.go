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

// A Writer is an io.WriteCloser.
// Writes to a Writer are compressed and written to w.
type Writer struct {
	w     io.Writer
	level int
	enc   *blockEncoder

	wroteHeader bool
	closed      bool
}

func NewWriter(w io.Writer) *Writer {
	z, _ := NewWriterLevel(w, DefaultCompression)
	return z
}

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

func (w *Writer) WriteFloat(f float64) error {
	return w.writeFloat64(f)
}

func (w *Writer) Flush() error {
	if err := w.ensureHeader(); err != nil {
		return err
	}
	return w.enc.flush()
}

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
