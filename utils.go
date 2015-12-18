package fpc

import (
	"bytes"
	"strconv"
	"strings"
)

// Utilities for using binary string representations of values

func bytes2binstr(bs []byte) string {
	var ss []string
	for _, b := range bs {
		s := strconv.FormatUint(uint64(b), 2)
		for len(s) < 8 {
			s = "0" + s
		}
		ss = append(ss, s)
	}
	return strings.Join(ss, " ")
}

func binstr2bytes(s string) []byte {
	s = strings.Replace(s, " ", "", -1)

	var bs []byte
	for len(s) > 0 {
		end := 8
		if 8 > len(s) {
			end = len(s)
		}
		val, err := strconv.ParseUint(s[:end], 2, 8)
		if err != nil {
			panic(err)
		}
		bs = append(bs, byte(val))
		s = s[end:]
	}
	return bs
}

func u642binstr(x uint64) string {
	s := strconv.FormatUint(x, 2)
	for len(s) < 64 {
		s = "0" + s
	}
	return insertNth(s, 8)
}
func u82binstr(x uint8) string {
	s := strconv.FormatUint(uint64(x), 2)
	for len(s) < 8 {
		s = "0" + s
	}
	return s
}

func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n-1 && i != len(s)-1 {
			buffer.WriteRune(' ')
		}
	}
	return buffer.String()
}

func binstr2u64(s string) uint64 {
	s = strings.Replace(s, " ", "", -1)

	val, err := strconv.ParseUint(s, 2, 64)
	if err != nil {
		panic(err)
	}
	return val
}

func binstr2u8(s string) uint8 {
	s = strings.Replace(s, " ", "", -1)

	val, err := strconv.ParseUint(s, 2, 8)
	if err != nil {
		panic(err)
	}
	return uint8(val)
}

func binstr2byte(s string) byte {
	return byte(binstr2u8(s))
}
