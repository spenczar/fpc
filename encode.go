package fpc

const encoderBuffer = 1024 // initial size for encoder buffer

func countLeadingZeroes(bits uint64) uint8 {
	if bits == 0 {
		return 64
	}
	var i uint8
	// 'while top bit is zero'
	for i = 0; bits&0x8000000000000000 == 0; i++ {
		bits <<= 1
	}
	return i
}

type encoder struct {
	bitPos int
	buf    []byte

	onHalfByte bool
}

func newEncoder() *encoder {
	return &encoder{
		bitPos:     0,
		buf:        make([]byte, encoderBuffer),
		onHalfByte: false,
	}
}

// bytes returns the fully-written bytes of the encoder
func (e *encoder) bytes() []byte {
	end := e.bitPos / 8
	if e.bitPos%8 == 4 {
		end += 1
	}
	return e.buf[:end]
}

// grow expands the internal buffer by max(atLeast, len(buffer))
func (e *encoder) grow(atLeast int) {
	growBy := len(e.buf)
	if atLeast > growBy {
		growBy = atLeast
	}
	e.buf = append(e.buf, make([]byte, growBy)...)
}

// encode writes a 64-bit value to the encoder's internal buffer
//
// TODO: profile this, it's not optimized at all
func (e *encoder) encode(diff uint64) []byte {
	nZeroBits := countLeadingZeroes(diff)

	nZeroNibbles := nZeroBits / 4
	// We only have 4 bits to store the number of zero nibbles, which means we
	// can only store numbers up to 15. It's possible that the entire diff is
	// zeroes, in which case nZeroNibbles is 16; if that's the case, then we
	// just store '15' in the first 4 bits and then a fully-zeroed nibble
	// afterwards, so '11110000'.
	if nZeroNibbles == 16 {
		nZeroNibbles = 15
	}

	// The input is a 64-bit value; after removing 4 * nZeroNibbles bits, we
	// have the number of data bits to store.
	nDataBits := 64 - nZeroNibbles*4

	// We need 4 bits to store the number of zeroes, plus data
	nBits := int(4 + nDataBits)

	// Compute the buffer indices to write to, which are measured in bytes, not
	// bits.
	start, end := e.bitPos/8, (e.bitPos+nBits)/8
	// If a fractional byte is left over, we'll need that, too. This will
	// happen if we're writing full bytes but start on a fractional byte, or if
	// we write fractional bytes and start on a full byte.
	if (nBits%8 == 0 && e.bitPos%8 == 4) ||
		(nBits%8 == 4 && e.bitPos%8 == 0) {
		end += 1
	}
	// If necessary, expand the internal buffer
	if end >= len(e.buf) {
		e.grow(end - start)
	}
	// Work on just a cleanly-isolated subslice of the data
	buf := e.buf[start:end]

	// Now, write the 4-bit prefix which encodes the number of leading zeroes
	// in diff.
	//
	// If a previous write left us in the middle of a byte, then we should
	// write to the lower 4 bits of the byte. Otherwise, write to the top 4
	// bits, and write the first 4 bits of the data bits to the bottom 4 bits.
	bitsWritten := 0
	if e.bitPos%8 == 4 {
		// in middle of byte
		buf[0] |= byte(nZeroNibbles)
		// Slide the zero nibbles off
		diff = diff << (nZeroNibbles * 4)
		// we've only written the 4-bit zero nibble data
		bitsWritten += 4
	} else {
		buf[0] |= byte(nZeroNibbles) << 4
		// slide the zero nibbles off
		diff = diff << (nZeroNibbles * 4)
		// write the 4 highest bits of our 64 bit value by sliding it over by
		// 60 bits
		buf[0] |= byte(diff >> 60)
		// slide off the written data
		diff = diff << 4
		// We've written 4 bits of zero-data, and 4 bits of real data
		bitsWritten += 8
	}

	// Write the rest of the data bits in 8-bit blocks
	for i := 1; bitsWritten < nBits; i, bitsWritten = i+1, bitsWritten+8 {
		// Sliding the 64-bit diff over by 56 bits leaves the top 8 bits
		buf[i] = byte(diff >> 56)
		// Slide off the written data
		diff = diff << 8
	}

	// There could be a 4-bit chunk of data remaining to write, because we've
	// been chomping off 8-bit blocks, but we stripped away a multiiple-of-four
	// number of zeroes. If this happens, then bitsWritten will be incremented
	// past nBits in the above loop.
	if bitsWritten > nBits {
		// write to the highest 4 bites of the last byte
		buf[len(buf)-1] |= byte(diff >> 60)
	}

	// Update our position in buf
	e.bitPos = e.bitPos + nBits
	return buf
}
