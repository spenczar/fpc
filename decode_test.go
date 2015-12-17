package fpc

import (
	"reflect"
	"testing"
)

func TestDecodePrefix(t *testing.T) {
	type output struct {
		n1, n2 uint8
		p1, p2 predictorClass
	}
	testcases := []struct {
		in   byte
		want output
	}{
		{
			in: binstr2byte("11101110"),
			want: output{
				n1: 7,
				n2: 7,
				p1: 0,
				p2: 0,
			},
		},
		{
			in: binstr2byte("11111110"),
			want: output{
				n1: 7,
				n2: 7,
				p1: 1,
				p2: 0,
			},
		},
		{
			in: binstr2byte("01001111"),
			want: output{
				n1: 2,
				n2: 7,
				p1: 0,
				p2: 1,
			},
		},
	}
	for i, tc := range testcases {
		var have output
		have.n1, have.n2, have.p1, have.p2 = decodePrefix(tc.in)
		if !reflect.DeepEqual(have, tc.want) {
			t.Errorf("decodePrefix test=%d  have=%+v  want=%+v", i, have, tc.want)
		}
	}
}
