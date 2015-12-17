package fpc

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func binstr2u64(s string) uint64 {

	s = strings.Join(strings.Split(s, " "), "")

	val, err := strconv.ParseUint(s, 2, 64)
	if err != nil {
		panic(err)
	}
	return val
}

// func bytes2binstr(bs []byte) string {
// 	var ss []string
// 	for _, b := range bs {
// 		s := strconv.FormatUint(uint64(b), 2)
// 		for len(s) < 8 {
// 			s = "0" + s
// 		}
// 		ss = append(ss, s)
// 	}
// 	return strings.Join(ss, " ")
// }

// func u642binstr(x uint64) string {
// 	s := strconv.FormatUint(x, 2)
// 	for len(s) < 64 {
// 		s = "0" + s
// 	}
// 	for i := 0; i < 64; i+= 8 {
// 		s = append(s, '')
// 		copy(s[i+1:], s[i:])
// 		s[i] = ' '
// 	}
// 	return s
// }

func TestCountLeadingZeroes(t *testing.T) {
	testcases := []struct {
		in   string
		want uint8
	}{
		{"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000", 64},
		{"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001", 63},
		{"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000010", 62},
		{"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000011", 62},
		{"00000000 00000000 00000000 00000000  10000000 00000000 00000000 00000011", 32},
		{"00000000 00000000 00000000 00000000  01111111 11111111 11111111 11111111", 33},
		{"00000000 00000000 00000000 00000000  11111111 11111111 11111111 11111111", 32},
		{"01111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111", 1},
		{"11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111", 0},
	}

	for _, tc := range testcases {
		have := countLeadingZeroes(binstr2u64(tc.in))
		if have != tc.want {
			t.Errorf("countLeadingZeroes  in=%d  have=%d  want=%d", tc.in, have, tc.want)
		}
	}
}

func BenchmarkLeadingZeroes(b *testing.B) {
	vals := make([]uint64, b.N, b.N)
	for i := range vals {
		vals[i] = math.Float64bits(rand.Float64())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countLeadingZeroes(vals[i])
	}
}

func TestEncodeDiff(t *testing.T) {
	testcases := []struct {
		in   string
		want string
	}{
		{
			// 0: there are 15 leading zero nibbles ('1111'), then '0000'.
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000000",
			want: "11110000",
		},
		{
			// 1: there are 15 leading zero nibbles, then '0001'.
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			want: "11110001",
		},
		{
			// 2: there are 15 leading zero nibbles, then '0011'.
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000011",
			want: "11110011",
		},
		{
			// 5: there are 15 leading zero nibbles, then '0101'.
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000101",
			want: "11110101",
		},
		{
			// 8: there are 15 leading zero nibbles, then '1000'.
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00001000",
			want: "11111000",
		},
		{
			// 24: there are 14 leading zero nibbles ('1110'), then '00011000' followed by 0s
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			want: "11100001 10000000",
		},
		{
			// 256: there are 13 leading zero nibbles ('1101')
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000001 00000000",
			want: "11010001 00000000",
		},
		{
			// 257: there are 13 leading zero nibbles ('1101')
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00000001 00000001",
			want: "11010001 00000001",
		},
		{
			// 4353: there are 12 leading zero nibbles ('1100')
			in:   "00000000 00000000 00000000 00000000  00000000 00000000 00010001 00000001",
			want: "11000001 00010000 00010000",
		},
		{
			// 8 leading zero nibbles
			in:   "00000000 00000000 00000000 00000000  11111111 11111111 11111111 11111111",
			want: "10001111 11111111 11111111 11111111 11110000",
		},
		{
			// 0 leading zero nibbles
			in:   "11111111 11111111 11111111 11111111  11111111 11111111 11111111 11111111",
			want: "00001111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11110000",
		},
	}
	for _, tc := range testcases {
		e := newEncoder()
		have := bytes2binstr(e.encode(binstr2u64(tc.in)))
		if have != tc.want {
			t.Errorf("encodeDiff error")
			t.Logf("  in=%s", tc.in)
			t.Logf("have=%s", have)
			t.Logf("want=%s", tc.want)
		}
	}
}

func TestEncoder(t *testing.T) {
	testcases := []struct {
		in   []string
		want string
	}{
		{
			in: []string{
				// Two 8-bit values
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000011",
			},
			want: "11110001 11110011",
		},
		{
			in: []string{
				// Two 12-bit values
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			},
			want: "11100001 10001110 00011000",
		},
		{
			in: []string{
				// Three 12-bit values
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			},
			want: "11100001 10001110 00011000 11100001 10000000",
		},
		{
			in: []string{
				// Four 12-bit values
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			},
			want: "11100001 10001110 00011000 11100001 10001110 00011000",
		},
		{
			in: []string{
				// One 8-bit value, then one 12-bit value
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			},
			want: "11110001 11100001 10000000",
		},
		{
			in: []string{
				// One 12-bit value, then one 8-bit value
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
			},
			want: "11100001 10001111 00010000",
		},
		{
			in: []string{
				// One 12-bit value, then one 8-bit value, then back to a 12-bit
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000001",
				"00000000 00000000 00000000 00000000  00000000 00000000 00000000 00011000",
			},
			want: "11100001 10001111 00011110 00011000",
		},
	}
	for i, tc := range testcases {
		e := newEncoder()
		for _, in := range tc.in {
			e.encode(binstr2u64(in))
		}
		have := bytes2binstr(e.bytes())
		if have != tc.want {
			t.Errorf("encodeDiff error case %d", i)
			for i, in := range tc.in {
				t.Logf(" in%d=%s", i, in)
			}
			t.Logf("have=%s", have)
			t.Logf("want=%s", tc.want)
		}
	}
}

func BenchmarkEncodeDiff(b *testing.B) {
	vals := make([]uint64, b.N, b.N)
	for i := range vals {
		vals[i] = math.Float64bits(rand.ExpFloat64())
	}

	e := newEncoder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.encode(vals[i])
	}
}
