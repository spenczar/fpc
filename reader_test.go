package fpc

import (
	"bytes"
	"testing"
)

func TestReader(t *testing.T) {
	r := NewReader(new(bytes.Buffer))

	for _, tc := range refTests {
		comp := bytes.NewBuffer(tc.compressed)
		r.Reset(comp)

		var err error
		have := make([]float64, len(tc.uncompressed))
		for i := range have {
			have[i], err = r.ReadFloat()
			tc.AssertNoError(t, err, "ReadFloat")
		}
		tc.AssertEqual(t, have, tc.uncompressed, "Reader")
	}
}
