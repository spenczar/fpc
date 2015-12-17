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
				n1: 8,
				n2: 8,
				p1: 0,
				p2: 0,
			},
		},
		{
			in: binstr2byte("11111110"),
			want: output{
				n1: 8,
				n2: 8,
				p1: 1,
				p2: 0,
			},
		},
		{
			in: binstr2byte("01001111"),
			want: output{
				n1: 2,
				n2: 8,
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

func TestDecodeOne(t *testing.T) {
	type output struct {
		n      int
		v1, v2 uint64
	}
	testcases := []struct {
		in   string
		want output
	}{
		{
			in: "11101110",
			want: output{
				n:  1,
				v1: 0,
				v2: 0,
			},
		},
		{
			in: "11101100 00000001",
			want: output{
				n:  2,
				v1: 0,
				v2: 1,
			},
		},
		{
			in: "11001100 00000001 00000001",
			want: output{
				n:  3,
				v1: 1,
				v2: 1,
			},
		},
		{
			in: "01101100 11111111 11111111 00000000  00000000 00000000 00000001",
			want: output{
				n:  7,
				v1: 1099494850560,
				v2: 1,
			},
		},
	}
	for i, tc := range testcases {
		var have output
		have.n, have.v1, have.v2 = decodeOne(binstr2bytes(tc.in))
		if !reflect.DeepEqual(have, tc.want) {
			t.Errorf("decodeOne test=%d  have=%+v  want=%+v", i, have, tc.want)
		}
	}
}
