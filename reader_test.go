package fpc

import (
	"bytes"
	"testing"
)

func TestReader(t *testing.T) {
	for _, tc := range refTests {
		comp := bytes.NewBuffer(tc.compressed)

		r := NewReader(comp)

		have := make([]float64, len(tc.uncompressed))
		_, err := r.ReadFloats(have)
		tc.AssertNoError(t, err, "ReadFloats")
		tc.AssertEqual(t, have, tc.uncompressed, "Reader")
	}
}
