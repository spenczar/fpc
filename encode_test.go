package fpc

import (
	"bytes"
	"reflect"
	"testing"
)

func TestPrefixCode(t *testing.T) {
	type output struct {
		code       uint8
		nZeroBytes uint64
	}
	testcases := []struct {
		in   string
		p    predictorClass
		want output
	}{
		{
			in: "11111111 00000000 00000000 00000000  00000000 00000000 00000000 10000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0000"),
				nZeroBytes: 0,
			},
		},
		{
			in: "11111111 00000000 00000000 00000000  00000000 00000000 00000000 10000000",
			p:  dfcmPredictor,
			want: output{
				code:       binstr2u8("0001"),
				nZeroBytes: 0,
			},
		},
		{
			in: "00000000 11111111 00000000 00000000  00000000 00000000 00000000 10000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0010"),
				nZeroBytes: 1,
			},
		},
		{
			in: "00000000 00000000 00000000 11111111  00000000 00000000 00000000 10000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0110"),
				nZeroBytes: 3,
			},
		},
		{
			in: "00000000 00000000 00000000 00000000  11111111 00000000 00000000 10000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("0110"),
				nZeroBytes: 3,
			},
		},
		{
			in: "00000000 00000000 00000000 00000000  00000000 11111111 00000000 10000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("1000"),
				nZeroBytes: 5,
			},
		},
		{
			in: "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000",
			p:  fcmPredictor,
			want: output{
				code:       binstr2u8("1110"),
				nZeroBytes: 8,
			},
		},
	}

	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		have := output{}
		have.code, have.nZeroBytes = e.prefixCode(binstr2u64(tc.in), tc.p)
		if !reflect.DeepEqual(tc.want, have) {
			t.Errorf("prefixCode test=%d  have.code=%s have.n=%d want.code=%s want.n=%d", i,
				u82binstr(have.code), have.nZeroBytes,
				u82binstr(tc.want.code), tc.want.nZeroBytes)
		}
	}
}

func TestEncodeNonzero(t *testing.T) {
	type input struct {
		val   string
		nZero uint64
	}
	testcases := []struct {
		in   input
		want string
	}{
		{
			in: input{
				val:   "11111111 00000000 00000000 00000000  00000000 00000000 00000000 10000000",
				nZero: 0,
			},
			want: "11111111 00000000 00000000 00000000  00000000 00000000 00000000 10000000",
		},
		{
			in: input{
				val:   "00000000 11111111 00000000 00000000  00000000 00000000 00000000 10000000",
				nZero: 1,
			},
			want: "11111111 00000000 00000000  00000000 00000000 00000000 10000000",
		},
		{
			in: input{
				val:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 10101010",
				nZero: 7,
			},
			want: "10101010",
		},
		{
			in: input{
				val:   "00000000 00000000 00000000 00000000  10101010 00000000 00000000 10101010",
				nZero: 3,
			},
			want: "00000000 10101010 00000000 00000000  10101010",
		},
	}

	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		have := make([]byte, 8-tc.in.nZero)
		e.encodeNonzero(binstr2u64(tc.in.val), tc.in.nZero, have)
		if !bytes.Equal(have, binstr2bytes(tc.want)) {
			t.Errorf("encodeNonzero test=%d", i)
			t.Logf("have=%s", bytes2binstr(have))
			t.Logf("want=%s", tc.want)
		}
	}
}

func TestPairEncode(t *testing.T) {
	testcases := []struct {
		v1, v2 string
		want   string
	}{
		{
			v1:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000",
			v2:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000",
			want: "11101110",
		},
		{
			v1:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			v2:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			want: "11001100 00000001 00000001",
		},
		{
			v1:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000",
			v2:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			want: "11101100 00000001",
		},
		{
			v1:   "00000000 00000000 00000000 11111111  11111111 00000000 00000000 00000000",
			v2:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			want: "01101100 11111111 11111111 00000000  00000000 00000000 00000001",
		},
		{
			v1: "11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111",
			v2: "11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111",
			want: "00000000 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111 " +
				"11111111",
		},
	}
	for i, tc := range testcases {
		e := newEncoder(defaultCompression)
		e.fcm = &mockPredictor{0}
		e.dfcm = &mockPredictor{0}
		have := e.encode(binstr2u64(tc.v1), binstr2u64(tc.v2))
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
