package fpc

import (
	"bytes"
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
	e := newBlockEncoder(w, DefaultCompression)
	e.enc.fcm = &mockPredictor{0xFABF}
	e.enc.dfcm = &mockPredictor{0xFABF}
	b.SetBytes(8)
	for i := 0; i < b.N; i++ {
		e.encodeFloat(0xFAFF * float64(i))
	}
}

func BenchmarkLeadingZeroBytes(b *testing.B) {
	b.SetBytes(8)
	for i := 0; i < b.N; i++ {
		clzBytes(uint64(i * 0xDEADBEEF))
	}
}

func BenchmarkPairEncode(b *testing.B) {
	e := newEncoder(DefaultCompression)
	e.fcm = &mockPredictor{0xFABF}
	e.dfcm = &mockPredictor{0xFABF}
	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.encode(0xFAFF*uint64(i), 0x1234*uint64(i))
	}
}

func BenchmarkEncodeNonzero(b *testing.B) {
	e := newEncoder(DefaultCompression)
	buf := make([]byte, 8)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.encodeNonzero(uint64(i), uint8(i)%8, buf)
	}
}

func BenchmarkComputeDiff(b *testing.B) {
	e := newEncoder(DefaultCompression)
	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.computeDiff(uint64(i))
	}
}

func BenchmarkFCM(b *testing.B) {
	fcm := newFCM(1 << DefaultCompression)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fcm.predict()
		fcm.update(uint64(i))
	}
}

func BenchmarkDFCM(b *testing.B) {
	dfcm := newDFCM(1 << DefaultCompression)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfcm.predict()
		dfcm.update(uint64(i))
	}
}

var benchcase = reftestcase{
	comp:         3,
	uncompressed: []float64{1e-05, 0.0001, 0.001, 0.01, 0.1, 1, 100, 1000, 10000, 100000},
	compressed: []byte{
		0x03, 0x0a, 0x00, 0x00, 0x53, 0x00, 0x00, 0x77,
		0xee, 0xee, 0xee, 0xee, 0xf1, 0x68, 0xe3, 0x88,
		0xb5, 0xf8, 0xe4, 0x3e, 0x2d, 0x43, 0x1c, 0xeb,
		0xe2, 0x36, 0x1a, 0x3f, 0xd1, 0xea, 0xed, 0x39,
		0xaf, 0x54, 0x4a, 0x87, 0xbd, 0x5f, 0x95, 0xac,
		0x18, 0xd4, 0xe1, 0x8d, 0x37, 0xde, 0x78, 0xe3,
		0x3d, 0x69, 0x00, 0x6f, 0x81, 0x04, 0xc5, 0x1f,
		0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x7f, 0x3c,
		0xda, 0x38, 0x62, 0x2d, 0x7e, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x08, 0x06, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xba, 0x0f},
}

func BenchmarkReader(b *testing.B) {
	b.SetBytes(int64(len(benchcase.compressed)))
	in := bytes.NewBuffer(benchcase.compressed)
	out := make([]float64, len(benchcase.uncompressed))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(in)
		r.ReadFloats(out)
		in.Reset()
	}
}

func BenchmarkWriter(b *testing.B) {
	b.SetBytes(int64(len(benchcase.uncompressed) * 8))
	w, _ := NewWriterLevel(ioutil.Discard, int(benchcase.comp))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteFloat(benchcase.uncompressed[i%len(benchcase.uncompressed)])
	}
}
