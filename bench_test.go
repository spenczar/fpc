package fpc

import (
	"math"
	"math/rand"
	"testing"
)

func generateValues(n int) []uint64 {
	vals := make([]uint64, n)
	// generate up to 1M random values
	for i := range vals {
		vals[i] = math.Float64bits(rand.ExpFloat64())
	}
	return vals
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func BenchmarkLeadingZeroBytes(b *testing.B) {
	n := min(b.N, 1e6)
	vals := generateValues(n)
	b.SetBytes(8)
	b.ResetTimer()

	v := vals[2%b.N]
	for i := 0; i < b.N; i++ {
		clzBytes(v)
	}
}

func BenchmarkPairEncode(b *testing.B) {
	e := newEncoder()

	n := min(b.N, 1e6)
	vals := generateValues(n)
	b.SetBytes(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.encode(vals[i%n], vals[i%n])
	}
}
