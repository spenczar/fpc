package fpc

import "math"

const decoderBuffer = 1024

type decoder struct {
	vals []float64

	fcm  predictor
	dfcm predictor

	nRecords, nBytes           int
	nRecordsTotal, nBytesTotal int
}

func newDecoder(compression uint) *decoder {
	tableSize := uint(1 << compression)
	return &decoder{
		fcm:  newFCM(tableSize),
		dfcm: newDFCM(tableSize),
		vals: make([]float64, 0),
	}
}

func (d *decoder) decodeBlock(block []byte) {
	if len(block) == 0 {
		return
	}
	// Decode header
	bh := decodeBlockHeader(block)
	d.nRecordsTotal = int(bh.nRecords)
	d.nBytesTotal = int(bh.nBytes)

	// Allocate space for the records
	headers := make([]header, d.nRecordsTotal)
	d.vals = append(d.vals, make([]float64, 0, d.nRecordsTotal)...)

	// Advance through the block past the header
	block = block[blockHeaderSize:]

	// Read the record headers. Each byte contains two headers.
	for i := 0; i < d.nRecordsTotal; i += 2 {
		headers[i], headers[i+1] = decodeHeaders(block[i/2])
	}
	// Advance past the record headers.
	block = block[d.nRecordsTotal/2:]
	// Decode the actual values
	var (
		val  uint64
		pred uint64
		h    header
	)
	for i := 0; i < d.nRecordsTotal; i += 1 {
		h = headers[i]

		val = decodeData(block[:h.len])

		// XOR with the predictions to get back the true values
		if h.pType == fcmPredictor {
			pred = d.fcm.predict()
		} else {
			pred = d.dfcm.predict()
		}
		val = pred ^ val
		d.vals = append(d.vals, math.Float64frombits(val))
		d.fcm.update(val)
		d.dfcm.update(val)
		// Advance through the block
		block = block[h.len:]
	}
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
