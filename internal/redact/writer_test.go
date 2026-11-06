package redact

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestRedactionAcrossEveryChunkBoundary(t *testing.T) {
	input := "start token-long then token and ünicode-secret end"
	want := "start [REDACTED] then [REDACTED] and [REDACTED] end"
	for size := 1; size <= len(input); size++ {
		var output bytes.Buffer
		writer := New(&output, [][]byte{[]byte("token"), []byte("token-long"), []byte("ünicode-secret"), {}})
		for offset := 0; offset < len(input); offset += size {
			end := min(offset+size, len(input))
			if _, err := writer.Write([]byte(input[offset:end])); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if output.String() != want {
			t.Fatalf("chunk %d: %q", size, output.String())
		}
	}
}

func TestNonsecretPromptIsNotBuffered(t *testing.T) {
	var output bytes.Buffer
	writer := New(&output, [][]byte{[]byte("long-secret-value")})
	if _, err := writer.Write([]byte("Continue? ")); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Continue? " {
		t.Fatal("interactive prompt was held back")
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}

type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 0, nil }

func TestWriterErrors(t *testing.T) {
	writer := New(shortWriter{}, nil)
	if _, err := writer.Write([]byte("output")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("error = %v", err)
	}
	_ = writer.Close()
}
