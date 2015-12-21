package fpc

import (
	"reflect"
	"testing"
)

func TestDecodeHeader(t *testing.T) {
	type output struct {
		n1, n2 uint8
		p1, p2 predictorClass
	}
	testcases := []struct {
		in   byte
		want pairHeader
	}{
		{
			in: binstr2byte("01110111"),
			want: pairHeader{
				h1: header{
					len:   8,
					pType: 0,
				},
				h2: header{
					len:   8,
					pType: 0,
				},
			},
		},
		{
			in: binstr2byte("11110111"),
			want: pairHeader{
				h1: header{
					len:   8,
					pType: 1,
				},
				h2: header{
					len:   8,
					pType: 0,
				},
			},
		},
		{
			in: binstr2byte("00101111"),
			want: pairHeader{
				h1: header{
					len:   2,
					pType: 0,
				},
				h2: header{
					len:   8,
					pType: 1,
				},
			},
		},
	}
	for i, tc := range testcases {
		var have pairHeader
		have.h1, have.h2 = decodeHeaders(tc.in)
		if !reflect.DeepEqual(have, tc.want) {
			t.Errorf("decodePrefix test=%d  have=%+v  want=%+v", i, have, tc.want)
		}
	}
}

func TestDecodeBlockHeader(t *testing.T) {

	type result struct {
		nRec   int
		nBytes int
	}
	testcases := []struct {
		in   []byte
		want result
	}{
		{
			in: []byte{0x00, 0x80, 0x00, 0xb6, 0x35, 0x02},
			want: result{
				nRec:   32768,
				nBytes: 144822,
			},
		},
		{
			in: []byte{0x00, 0x80, 0x00, 0xc2, 0x43, 0x00},
			want: result{
				nRec:   32768,
				nBytes: 17346,
			},
		},
	}
	for i, tc := range testcases {
		var have result
		have.nRec, have.nBytes = decodeBlockHeader(tc.in)
		if !reflect.DeepEqual(have, tc.want) {
			t.Errorf("decodeBlockHeader test=%d  have=%+v  want=%+v", i, have, tc.want)
		}
	}
}
