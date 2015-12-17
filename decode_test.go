package fpc

import (
	"math"
	"testing"
)

func TestDecodeOne(t *testing.T) {
	testcases := []struct {
		in   string
		want uint64
	}{
		{
			in:   "11110000",
			want: 0,
		},
		{
			in:   "11110001",
			want: 1,
		},
		{
			in:   "11110010",
			want: 2,
		},
		{
			in:   "11100001 10000000",
			want: 24,
		},
		{
			in:   "11010001 00000001",
			want: 257,
		},
		{
			in:   "11000001 00010000 00010000",
			want: 4353,
		},
		{
			in:   "00001111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11110000",
			want: math.MaxUint64,
		},
	}

	for _, tc := range testcases {
		d := newDecoder()
		have := d.decodeOneBitwise(binstr2bytes(tc.in))
		if have != tc.want {
			t.Errorf("encodeDiff error")
			t.Logf("  in=%v", tc.in)
			t.Logf("have=%v", have)
			t.Logf("want=%v", tc.want)
		}
	}

}
