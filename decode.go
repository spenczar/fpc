package fpc

const decoderBuffer = 1024

type decoder struct {
	vals []uint64
}

func newDecoder() *decoder {
	return &decoder{
		vals: make([]uint64, decoderBuffer),
	}
}

func decodePrefix(b byte) (nzero1, nzero2 uint8, p1, p2 predictorClass) {
	nzero1, nzero2 = uint8((b&0xE0)>>5), uint8(b&0x0E>>1)
	// See the comment in encoder.prefixCode.
	if nzero1 >= 4 {
		nzero1 += 1
	}
	if nzero2 >= 4 {
		nzero2 += 1
	}
	p1, p2 = predictorClass((b&0x10)>>4), predictorClass(b&0x01)
	return
}

func decodeOne(b []byte) (nRead int, v1, v2 uint64) {
	// Pull out prefix bits
	nz1, nz2, _, _ := decodePrefix(b[0])

	start, end := 1, 1+8-int(nz1)
	v1 = decodeData(b[start:end])

	start, end = end, end+8-int(nz2)
	v2 = decodeData(b[start:end])

	return end, v1, v2
}

func decodeData(bs []byte) (v uint64) {
	last := len(bs) - 1
	for i, b := range bs {
		v |= uint64(b)
		if i == last {
			break
		}
		v <<= 8
	}
	return v
}

type blockHeader struct {
	nRecords, nBytes uint32
}

func decodeFirstBlock(bs []byte) (compression int, h blockHeader) {
	// First byte encodes size of hash tables
	h = decodeBlockHeader(bs[1:])
	return int(bs[0]), h
}

func decodeBlockHeader(bs []byte) blockHeader {
	var h blockHeader
	// First three bytes encode the number of records
	h.nRecords = uint32(bs[2])
	h.nRecords = (h.nRecords << 8) | uint32(bs[1])
	h.nRecords = (h.nRecords << 8) | uint32(bs[0])

	// remaining 3 encode the number of bytes in the block
	h.nBytes = uint32(bs[5])
	h.nBytes = (h.nBytes << 8) | uint32(bs[4])
	h.nBytes = (h.nBytes << 8) | uint32(bs[3])

	return h
}
