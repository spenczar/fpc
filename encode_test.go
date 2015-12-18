package fpc

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestBlockEncoder(t *testing.T) {
	testcases := []struct {
		comp uint
		vals []float64
		want []byte
	}{
		{
			comp: 1,
			vals: []float64{1, 1},
			want: []byte{
				0x02, 0x00, 0x00, 0x0f, 0x00, 0x00, 0x70, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f,
			},
		},
	}
	for _, tc := range testcases {
		buf := new(bytes.Buffer)
		e := newBlockEncoder(buf, tc.comp)
		for _, v := range tc.vals {
			if err := e.encode(v); err != nil {
				t.Fatalf("encode err=%q", err)
			}
		}
		if err := e.flush(); err != nil {
			t.Fatalf("flush err=%q", err)
		}
		if have := buf.Bytes(); !bytes.Equal(have, tc.want) {
			t.Error("block encode")
			t.Logf("have=%#v", bytes2binstr(have))
			t.Logf("want=%#v", bytes2binstr(tc.want))
		}
	}
}

func TestPrefixCode(t *testing.T) {
	type output struct {
		code       uint8
		nZeroBytes uint64
	}
	testcases := []struct {
		in   []byte
		p    predictorClass
		want output
	}{
		{
			in: []byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0000"),
				nZeroBytes: 0,
			},
		},
		{
			in: []byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			p:  dfcmPredictor,
			want: output{
				code:       binstr2u8("1000"),
				nZeroBytes: 0,
			},
		},
		{
			in: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0001"),
				nZeroBytes: 1,
			},
		},
		{
			in: []byte{0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0011"),
				nZeroBytes: 3,
			},
		},
		{
			in: []byte{0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0011"),
				nZeroBytes: 3,
			},
		},
		{
			in: []byte{0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0100"),
				nZeroBytes: 5,
			},
		},
		{
			in: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0111"),
				nZeroBytes: 8,
			},
		},
	}

	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		have := output{}
		have.code, have.nZeroBytes = e.prefixCode(binary.LittleEndian.Uint64(tc.in), tc.p)
		if !reflect.DeepEqual(tc.want, have) {
			t.Errorf("prefixCode test=%d  have.code=%s have.n=%d want.code=%s want.n=%d", i,
				u82binstr(have.code), have.nZeroBytes,
				u82binstr(tc.want.code), tc.want.nZeroBytes)
		}
	}
}

func TestEncodeNonzero(t *testing.T) {
	type input struct {
		val   []byte
		nZero uint64
	}
	testcases := []struct {
		in   input
		want []byte
	}{
		{
			in: input{
				val:   []byte{0xFF, 0, 0, 0, 0, 0, 0, 0XFF},
				nZero: 0,
			},
			want: []byte{0xFF, 0, 0, 0, 0, 0, 0, 0XFF},
		},
		{
			in: input{
				val:   []byte{0xFF, 0, 0, 0, 0, 0, 0xFF, 0},
				nZero: 1,
			},
			want: []byte{0xFF, 0, 0, 0, 0, 0, 0xFF},
		},
		{
			in: input{
				val:   []byte{0xAA, 0, 0, 0, 0, 0, 0, 0},
				nZero: 7,
			},
			want: []byte{0xAA},
		},
		{
			in: input{
				val:   []byte{0xAA, 0, 0, 0, 0xAA, 0, 0, 0},
				nZero: 3,
			},
			want: []byte{0xAA, 0, 0, 0, 0xAA},
		},
	}

	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		have := make([]byte, 8-tc.in.nZero)
		e.encodeNonzero(byteOrder.Uint64(tc.in.val), tc.in.nZero, have)
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
		{
			v1:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
			v2:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
			want: "01110111",
		},
		{
			v1:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			want: "01100110 00000001 00000001",
		},
		{
			v1:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
			v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			want: "01110110 00000001",
		},
		{
			v1:   []byte{0, 0, 0, 0xFF, 0xFF, 0, 0, 0},
			v2:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
			want: "00110110 00000000 00000000 00000000  11111111 11111111 00000001",
		},
		{
			v1: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			v2: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			want: "00000000 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111",
		},
	}
	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		e.fcm = &mockPredictor{0}
		e.dfcm = &mockPredictor{0}
		have := e.encode(byteOrder.Uint64(tc.v1), byteOrder.Uint64(tc.v2))
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
