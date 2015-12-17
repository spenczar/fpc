package fpc

const encoderBuffer = 1024 // initial size for encoder buffer

type encoder struct {
	buf []byte
}

func newEncoder() *encoder {
	return &encoder{
		buf: make([]byte, 17),
	}
}

// encode a pair of values
func (e *encoder) encode(v1, v2 uint64) []byte {
	v1Prefix, v1NZero := e.prefixCode(v1)
	v2Prefix, v2NZero := e.prefixCode(v2)

	// First byte contains prefixes for the values
	e.buf[0] = v1Prefix<<4 | v2Prefix

	_, _ = v1NZero, v2NZero
	e.encodeNonzero(v1, v1NZero, e.buf[1:])
	e.encodeNonzero(v2, v2NZero, e.buf[1+(8-v1NZero):])
	return e.buf[:1+(8-v1NZero)+(8-v2NZero)]
}

func (e *encoder) encodeNonzero(v uint64, nZero uint64, into []byte) {
	// Starting with the first nonzero byte, copy v's data into the byte slice.
	//
	// Unrolling this loop into a switch speeds up the computation dramatically.

	switch nZero {
	case 0:
		into[0] = byte((v >> 56) & 0xFF)
		into[1] = byte((v >> 48) & 0xFF)
		into[2] = byte((v >> 40) & 0xFF)
		into[3] = byte((v >> 32) & 0xFF)
		into[4] = byte((v >> 24) & 0xFF)
		into[5] = byte((v >> 16) & 0xFF)
		into[6] = byte((v >> 8) & 0xFF)
		into[7] = byte(v & 0xFF)
	case 1:
		into[0] = byte((v >> 48) & 0xFF)
		into[1] = byte((v >> 40) & 0xFF)
		into[2] = byte((v >> 32) & 0xFF)
		into[3] = byte((v >> 24) & 0xFF)
		into[4] = byte((v >> 16) & 0xFF)
		into[5] = byte((v >> 8) & 0xFF)
		into[6] = byte(v & 0xFF)
	case 2:
		into[0] = byte((v >> 40) & 0xFF)
		into[1] = byte((v >> 32) & 0xFF)
		into[2] = byte((v >> 24) & 0xFF)
		into[3] = byte((v >> 16) & 0xFF)
		into[4] = byte((v >> 8) & 0xFF)
		into[5] = byte(v & 0xFF)
	case 3:
		into[0] = byte((v >> 32) & 0xFF)
		into[1] = byte((v >> 24) & 0xFF)
		into[2] = byte((v >> 16) & 0xFF)
		into[3] = byte((v >> 8) & 0xFF)
		into[4] = byte(v & 0xFF)
	case 4:
		into[0] = byte((v >> 24) & 0xFF)
		into[1] = byte((v >> 16) & 0xFF)
		into[2] = byte((v >> 8) & 0xFF)
		into[3] = byte(v & 0xFF)
	case 5:
		into[0] = byte((v >> 16) & 0xFF)
		into[1] = byte((v >> 8) & 0xFF)
		into[2] = byte(v & 0xFF)
	case 6:
		into[0] = byte((v >> 8) & 0xFF)
		into[1] = byte(v & 0xFF)
	case 7:
		into[0] = byte(v & 0xFF)
	}
}

// Compute 4-bit header for the value
func (e *encoder) prefixCode(v uint64) (code uint8, nZeroBytes uint64) {
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
	if e.choosePredictor(v) == fcmPredictor {
		return uint8(z<<1 | 0), zOrig
	} else {
		return uint8(z<<1 | 1), zOrig
	}
}

func (e *encoder) choosePredictor(v uint64) predictorClass {
	return fcmPredictor
}
