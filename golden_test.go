package fpc

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

const (
	goldenCompressedFilepath   = "golden/test.trace.fpc"
	goldenDecompressedFilepath = "golden/test_decompressed.data"
)

func TestGoldenCompress(t *testing.T) {
	// Test that compressing data reproduces the reference implementation's
	// compression.
	input, err := os.Open(goldenDecompressedFilepath)
	if err != nil {
		t.Fatalf("unable to load decompressed bytes: %v", err)
	}
	defer input.Close()

	buf := bytes.NewBuffer(nil)
	w, err := NewWriterLevel(buf, 20)
	if err != nil {
		t.Fatalf("unable to create writer: %v", err)
	}

	_, err = io.Copy(w, input)
	if err != nil {
		t.Fatalf("unable to write: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("unable to close: %v", err)
	}

	want, err := ioutil.ReadFile(goldenCompressedFilepath)
	if err != nil {
		t.Fatalf("unable to load golden compressed bytes: %v", err)
	}

	have := buf.Bytes()
	if !bytes.Equal(have, want) {
		t.Error("compressed data golden mismatch")
		t.Logf("len(have) = %d", len(have))
		t.Logf("len(want) = %d", len(want))
	}
}

func TestGoldenDecompress(t *testing.T) {
	// Test that decompressing data reproduces the reference implementation's
	// decompression.
	input, err := os.Open(goldenCompressedFilepath)
	if err != nil {
		t.Fatalf("unable to load compressed bytes: %v", err)
	}
	defer input.Close()

	r := NewReader(input)
	readBuf := make([]byte, 1024)
	haveBuf := bytes.NewBuffer(nil)
	for {
		n, err := r.Read(readBuf)
		if err == io.EOF {
			haveBuf.Write(readBuf[:n])
			break
		} else if err != nil {
			t.Fatalf("read error: %v", err)
		}
		haveBuf.Write(readBuf[:n])
	}

	want, err := ioutil.ReadFile(goldenDecompressedFilepath)
	if err != nil {
		t.Fatalf("unable to load golden decompressed bytes: %v", err)
	}

	have := haveBuf.Bytes()
	if !bytes.Equal(have, want) {
		t.Error("decompressed data golden mismatch")
		t.Logf("len(have) = %d", len(have))
		t.Logf("len(want) = %d", len(want))
	}
}
