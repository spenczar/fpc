package fpc

import (
	"reflect"
	"testing"
)

func TestDecodeBlock(t *testing.T) {
	for i, tc := range refTests {
		d := newDecoder(tc.comp)
		d.decodeBlock(tc.compressed[1:])
		if have := d.vals; !reflect.DeepEqual(have, tc.uncompressed) {
			t.Errorf("block decode test=%d", i)
			t.Logf("in=%#v", tc.compressed)
			t.Logf("have=%#v", have)
			t.Logf("want=%#v", tc.uncompressed)
		}
	}
}

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
	testcases := []struct {
		in   []byte
		want blockHeader
	}{
		{
			in: []byte{0x00, 0x80, 0x00, 0xb6, 0x35, 0x02},
			want: blockHeader{
				nRecords: 32768,
				nBytes:   144822,
			},
		},
		{
			in: []byte{0x00, 0x80, 0x00, 0xc2, 0x43, 0x00},
			want: blockHeader{
				nRecords: 32768,
				nBytes:   17346,
			},
		},
	}
	for i, tc := range testcases {
		have := decodeBlockHeader(tc.in)
		if !reflect.DeepEqual(have, tc.want) {
			t.Errorf("decodeBlockHeader test=%d  have=%+v  want=%+v", i, have, tc.want)
		}
	}
}
