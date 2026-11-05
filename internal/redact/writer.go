// Package redact masks exact credential values in subprocess output.
package redact

import (
	"bytes"
	"io"
	"sort"
)

// Writer retains trailing bytes to detect values split across writes.
// It handles one stream and does not detect arbitrary encodings or other IO.
type Writer struct {
	output  io.Writer
	secrets [][]byte
	pending []byte
}

func New(output io.Writer, secrets [][]byte) *Writer {
	w := &Writer{output: output}
	for _, secret := range secrets {
		if len(secret) == 0 {
			continue
		}
		w.secrets = append(w.secrets, bytes.Clone(secret))
	}
	sort.Slice(w.secrets, func(i, j int) bool { return len(w.secrets[i]) > len(w.secrets[j]) })
	return w
}

func (w *Writer) Write(p []byte) (int, error) {
	w.pending = append(w.pending, p...)
	if err := w.drain(false); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close flushes remaining output and clears owned credential buffers.
func (w *Writer) Close() error {
	defer func() {
		for _, secret := range w.secrets {
			clear(secret)
		}
		clear(w.pending)
		w.secrets = nil
		w.pending = nil
	}()
	return w.drain(true)
}

func (w *Writer) drain(final bool) error {
	var output []byte
	consumed := 0
	for consumed < len(w.pending) {
		matched := false
		incomplete := false
		for _, secret := range w.secrets {
			if !final && len(w.pending)-consumed < len(secret) && bytes.HasPrefix(secret, w.pending[consumed:]) {
				incomplete = true
				break
			}
			if bytes.HasPrefix(w.pending[consumed:], secret) {
				output = append(output, "[REDACTED]"...)
				consumed += len(secret)
				matched = true
				break
			}
		}
		if incomplete {
			break
		}
		if !matched {
			output = append(output, w.pending[consumed])
			consumed++
		}
	}
	if len(output) > 0 {
		n, err := w.output.Write(output)
		if err != nil {
			return err
		}
		if n != len(output) {
			return io.ErrShortWrite
		}
	}
	remaining := copy(w.pending, w.pending[consumed:])
	clear(w.pending[remaining:])
	w.pending = w.pending[:remaining]
	return nil
}
