// +build !amd64

package fpc

func clzBytes(val uint64) uint64 {
	if val == 0 {
		return 8
	}
	var i uint64
	// 'while top byte is zero'
	for i = 0; val&0xFF00000000000000 == 0; i++ {
		val <<= 8
	}
	return i
}
