package log

import (
	"bytes"
	"io"
	"sync"
)

type spacedWriter struct {
	w   io.Writer
	mu  sync.Mutex
	buf bytes.Buffer
}

func newSpacedWriter(w io.Writer) io.Writer {
	return &spacedWriter{w: w}
}

func (sw *spacedWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	n, err := sw.buf.Write(p)
	if err != nil {
		return n, err
	}

	for {
		b := sw.buf.Bytes()
		i := bytes.IndexByte(b, '\n')
		if i == -1 {
			break
		}

		line := make([]byte, i+1)
		copy(line, b[:i+1])
		sw.buf.Next(i + 1)

		if _, err := sw.w.Write(line); err != nil {
			return n, err
		}
		if _, err := sw.w.Write([]byte("\n")); err != nil {
			return n, err
		}
	}

	return n, nil
}
