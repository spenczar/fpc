package fpc

import "math/bits"

func clzBytes(val uint64) uint64 {
	x := bits.LeadingZeros64(val)
	return uint64(x / 8) // Integer division rounds down
}
