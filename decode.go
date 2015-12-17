package fpc

const decoderBuffer = 1024

type pairDecoder struct {
	vals []uint64
}

func newPairDecoder() *pairDecoder {
	return &pairDecoder{
		vals: make([]uint64, decoderBuffer),
	}
}

func decodePrefix(b byte) (nzero1, nzero2 uint8, p1, p2 predictorClass) {
	return uint8((b & 0xE0) >> 5), uint8(b & 0x0E >> 1), predictorClass((b & 0x10) >> 4), predictorClass(b & 0x01)
}

func (d *pairDecoder) decode(b []byte) (v1, v2 uint64) {
	// Pull out prefix bits
	return 0, 0
}
