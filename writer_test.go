package fpc

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	for _, tc := range refTests {
		have := bytes.NewBuffer(nil)
		w, err := NewWriterLevel(have, int(tc.comp))
		if err != nil {
			t.Fatalf("NewWriterLevel err=%q", err)
		}
		for _, f := range tc.uncompressed {
			err = w.WriteFloat(f)
			tc.AssertNoError(t, err, "WriteFloat")
		}
		err = w.Close()
		tc.AssertNoError(t, err, "Close")

		tc.AssertEqual(t, have.Bytes(), tc.compressed, "Writer")
	}
}

// Should not panic resetting a writer that initially had a nil underlying writer
func TestWriter_Reset_ExistingWriterNil(t *testing.T) {
	w, err := NewWriterLevel(nil, 1)
	if err != nil {
		t.Fatalf("NewWriterLevel err=%q", err)
	}

	err = w.Reset(new(bytes.Buffer), 1)
	if err != nil {
		t.Fatalf("Reset err=%q", err)
	}
}
