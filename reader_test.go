package fpc

import (
	"bytes"
	"testing"
)

func TestReader(t *testing.T) {
	for _, tc := range refTests {
		comp := bytes.NewBuffer(tc.compressed)

		r := NewReader(comp)

		var err error
		have := make([]float64, len(tc.uncompressed))
		for i := range have {
			have[i], err = r.ReadFloat()
			tc.AssertNoError(t, err, "ReadFloats")
		}
		tc.AssertEqual(t, have, tc.uncompressed, "Reader")
	}
}
