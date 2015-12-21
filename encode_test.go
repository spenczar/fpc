package fpc

import (
	"bytes"
	"testing"
)

func TestBlockEncoder(t *testing.T) {
	for i, tc := range refTests {
		buf := new(bytes.Buffer)
		e := newBlockEncoder(buf, tc.comp)
		for _, v := range tc.uncompressed {
			if err := e.encodeFloat(v); err != nil {
				t.Fatalf("encode err=%q", err)
			}
		}
		if err := e.flush(); err != nil {
			t.Fatalf("flush err=%q", err)
		}
		want := tc.compressed[1:] // strip leading byte which describes compression to use
		if have := buf.Bytes(); !bytes.Equal(have, want) {
			t.Errorf("block encode  test=%d", i)
			t.Logf("in.comp=%v", tc.comp)
			t.Logf("in.values=%#v", tc.compressed)
			t.Logf("have=%#v", bytes2binstr(have))
			t.Logf("want=%#v", bytes2binstr(want))
		}
	}
}

func TestEncodeHeader(t *testing.T) {
	testcases := []struct {
		h    pairHeader
		want byte
	}{
		{
			h: pairHeader{
				h1: header{
					len:   0,
					pType: 0,
				},
				h2: header{
					len:   0,
					pType: 0,
				},
			},
			want: 0,
		},
		{
			h: pairHeader{
				h1: header{
					len:   0,
					pType: 0,
				},
				h2: header{
					len:   1,
					pType: 0,
				},
			},
			want: 1,
		},
		{
			h: pairHeader{
				h1: header{
					len:   1,
					pType: 0,
				},
				h2: header{
					len:   0,
					pType: 0,
				},
			},
			want: 0x10,
		},
		{
			h: pairHeader{
				h1: header{
					len:   1,
					pType: 1,
				},
				h2: header{
					len:   1,
					pType: 1,
				},
			},
			want: 0x99,
		},
		{
			h: pairHeader{
				h1: header{
					len:   3,
					pType: 1,
				},
				h2: header{
					len:   7,
					pType: 1,
				},
			},
			want: 0xBE,
		},
	}
	for i, tc := range testcases {
		have := tc.h.encode()
		if have != tc.want {
			t.Errorf("header encoding err  test=%d", i)
			t.Logf("have=%#v", have)
			t.Logf("want=%#v", tc.want)
		}
	}
}

func TestEncodeNonzero(t *testing.T) {
	type input struct {
		val []byte
		len uint8
	}
	testcases := []struct {
		in   input
		want []byte
	}{
		{
			in: input{
				val: []byte{0xFF, 0, 0, 0, 0, 0, 0, 0XFF},
				len: 8,
			},
			want: []byte{0xFF, 0, 0, 0, 0, 0, 0, 0XFF},
		},
		{
			in: input{
				val: []byte{0xFF, 0, 0, 0, 0, 0, 0xFF, 0},
				len: 7,
			},
			want: []byte{0xFF, 0, 0, 0, 0, 0, 0xFF},
		},
		{
			in: input{
				val: []byte{0xAA, 0, 0, 0, 0, 0, 0, 0},
				len: 1,
			},
			want: []byte{0xAA},
		},
		{
			in: input{
				val: []byte{0xAA, 0, 0, 0, 0xAA, 0, 0, 0},
				len: 5,
			},
			want: []byte{0xAA, 0, 0, 0, 0xAA},
		},
	}

	for i, tc := range testcases {
		e := newEncoder(DefaultCompression)
		have := make([]byte, tc.in.len)
		e.encodeNonzero(byteOrder.Uint64(tc.in.val), tc.in.len, have)
		if !bytes.Equal(have, tc.want) {
			t.Errorf("encodeNonzero test=%d", i)
			t.Logf("have=%s", bytes2binstr(have))
			t.Logf("want=%s", tc.want)
		}
	}
}

func TestPairEncode(t *testing.T) {
	testcases := []struct {
		v1, v2 []byte
		want   string
	}{
		// {
		// 	v1:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
		// 	v2:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
		// 	want: "01110111",
		// },
		// {
		// 	v1:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
		// 	v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
		// 	want: "01100110 00000001 00000001",
		// },
		{
			v1:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
			v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			want: "00000001 00000001",
		},
		{
			v1:   []byte{0, 0, 0, 0xFF, 0xFF, 0, 0, 0},
			v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			want: "01000001 00000000 00000000 00000000  11111111 11111111 00000001",
		},
		{
			v1: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			v2: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			want: "01110111 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111",
		},
	}
	for i, tc := range testcases {
		e := newEncoder(DefaultCompression)
		e.fcm = &mockPredictor{0}
		e.dfcm = &mockPredictor{0}
		haveHeader, haveData := e.encode(byteOrder.Uint64(tc.v1), byteOrder.Uint64(tc.v2))
		have := append([]byte{haveHeader.encode()}, haveData...)
		if !bytes.Equal(have, binstr2bytes(tc.want)) {
			t.Errorf("encode test=%d", i)
			t.Logf("have=%s", bytes2binstr(have))
			t.Logf("want=%s", tc.want)
		}
	}
}

// mockPredictor always predicts the same value
type mockPredictor struct {
	val uint64
}

func (p *mockPredictor) predict() uint64 { return p.val }
func (p *mockPredictor) update(uint64)   {}
