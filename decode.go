package fpc

import "log"

const decoderBuffer = 1024 // initial size for decoder diff buffer

type decoder struct {
	diffs []uint64
}

func newDecoder() *decoder {
	return &decoder{
		diffs: make([]uint64, decoderBuffer),
	}
}

const (
	upper = 0xf0
	lower = 0x0f
)

func (d *decoder) decodeOne(b []byte) uint64 {
	// The top 4 bits indicate the number of leading zero nibbles in the
	// encoded float.
	nZeroNibbles := uint((b[0] & upper) >> 4)

	// Compute number of non-zero nibbles to read
	nDataNibbles := uint((64 / 4) - nZeroNibbles)
	nibblesRead := uint(0)

	var val uint64
	// Chomp off the bottom 4 bits of the first byte
	val = uint64(b[0]&lower) << (60 - nZeroNibbles*4)
	nibblesRead += 1
	log.Printf("val.0=%d", val)
	// Chomp off 8-byte chunks
	var i int

	log.Printf("read, total = %d, %d", nibblesRead, nDataNibbles)

	for i = 1; (nibblesRead + 1) < nDataNibbles; i, nibblesRead = i+1, nibblesRead+2 {
		add := uint64(b[i]) << (60 - (nibblesRead+1)*4 - nZeroNibbles*4)
		val |= add
		log.Printf("val.%d=%d", i, val)
		log.Printf("val.%d, added=%d", i, add)
		log.Printf("read, total = %d, %d", nibblesRead, nDataNibbles)
	}

	// If there is a remaining 4-bit chunk in the upper half of a byte, chomp
	// that off too
	if nibblesRead != nDataNibbles {
		log.Printf("last byte: %s", bytes2binstr([]byte{b[i]}))
		add := uint64(b[i]&upper) >> 4
		val |= add
		log.Printf("val.last=%d", val)
		log.Printf("val.last, added=%d", add)
	}

	return val
}

func (d *decoder) decodeOneBitwise(b []byte) uint64 {
	// The top 4 bits indicate the number of leading zero nibbles in the
	// encoded float.
	nZeroBits := uint((b[0]&upper)>>4) * 4

	// Compute number of non-zero nibbles to read
	nDataBits := uint(64 - nZeroBits)
	nBitsRead := uint(0)

	var val uint64
	// Chomp off the bottom 4 bits of the first byte
	val = uint64(b[0]&lower) << (60 - nZeroBits)
	nBitsRead += 4

	// Chomp off 8-byte chunks
	var i int
	for i = 1; (nBitsRead + 8) <= nDataBits; i, nBitsRead = i+1, nBitsRead+8 {
		add := uint64(b[i]) << (56 - nBitsRead - nZeroBits)
		val |= add
	}

	// If there is a remaining 4-bit chunk in the upper half of a byte, chomp
	// that off too
	if nBitsRead != nDataBits {
		add := uint64(b[i]&upper) >> 4
		val |= add
	}
	return val
}
