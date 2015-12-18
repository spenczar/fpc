package fpc

import (
	"io/ioutil"
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

func BenchmarkBlockEncode(b *testing.B) {
	w := ioutil.Discard
	e := newBlockEncoder(w, defaultCompression)
	e.enc.fcm = &mockPredictor{0xFABF}
	e.enc.dfcm = &mockPredictor{0xFABF}
	b.SetBytes(8)
	for i := 0; i < b.N; i++ {
		e.encode(0xFAFF * float64(i))
	}
}

func BenchmarkLeadingZeroBytes(b *testing.B) {
	b.SetBytes(8)
	for i := 0; i < b.N; i++ {
		clzBytes(uint64(i * 0xDEADBEEF))
	}
}

func BenchmarkPairEncode(b *testing.B) {
	e := newEncoder(defaultCompression)
	e.fcm = &mockPredictor{0xFABF}
	e.dfcm = &mockPredictor{0xFABF}
	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.encode(0xFAFF*uint64(i), 0x1234*uint64(i))
	}
}

func BenchmarkEncodeNonzero(b *testing.B) {
	e := newEncoder(defaultCompression)
	buf := make([]byte, 8)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.encodeNonzero(uint64(i), uint8(i)%8, buf)
	}
}

func BenchmarkComputeDiff(b *testing.B) {
	e := newEncoder(defaultCompression)
	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.computeDiff(uint64(i))
	}
}

func BenchmarkFCM(b *testing.B) {
	fcm := newFCM(1 << defaultCompression)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fcm.predict()
		fcm.update(0xFAFF)
	}
}

func BenchmarkDFCM(b *testing.B) {
	dfcm := newDFCM(1 << defaultCompression)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfcm.predict()
		dfcm.update(0xFAFF)
	}
}
