package fpc

func decodeBlockHeader(b []byte) (nRecords, nBytes int) {
	// First three bytes encode the number of records
	nRecordsUint := uint32(b[2])
	nRecordsUint = (nRecordsUint << 8) | uint32(b[1])
	nRecordsUint = (nRecordsUint << 8) | uint32(b[0])

	// remaining 3 encode the number of bytes in the block
	nBytesUint := uint32(b[5])
	nBytesUint = (nBytesUint << 8) | uint32(b[4])
	nBytesUint = (nBytesUint << 8) | uint32(b[3])

	return int(nRecordsUint), int(nBytesUint)
}

func decodeHeaders(b byte) (h1, h2 header) {
	h1 = header{
		len:   (b & 0x70) >> 4,
		pType: predictorClass((b & 0x80) >> 7),
	}
	h2 = header{
		len:   (b & 0x07),
		pType: predictorClass((b & 0x08) >> 3),
	}
	if h1.len >= 4 {
		h1.len += 1
	}
	if h2.len >= 4 {
		h2.len += 1
	}
	return h1, h2
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
