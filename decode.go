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
	// top bit and 5th bit are indicators for predictor classes
	nzero1, nzero2 = uint8((b&0x70)>>4), uint8(b&0x07)
	// See the comment in encoder.prefixCode.
	if nzero1 >= 4 {
		nzero1 += 1
	}
	if nzero2 >= 4 {
		nzero2 += 1
	}
	p1, p2 = predictorClass((b&0x80)>>7), predictorClass(b&0x08>>3)
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

func decodeData(b []byte) (v uint64) {
	// Decode b as a partial little-endian uint64
	switch len(b) {
	case 8:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32 |
			uint64(b[5])<<40 |
			uint64(b[6])<<48 |
			uint64(b[7])<<56)
	case 7:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32 |
			uint64(b[5])<<40 |
			uint64(b[6])<<48)
	case 6:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32 |
			uint64(b[5])<<40)
	case 5:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32)
	case 4:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24)
	case 3:
		v = (uint64(b[0]) |
			uint64(b[1])<<8 |
			uint64(b[2])<<16)
	case 2:
		v = (uint64(b[0]) |
			uint64(b[1])<<8)
	case 1:
		v = uint64(b[0])
		// case 0: leave v as 0
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
