package core

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ylSink buffers a Hall-Petch yield snapshot and flushes on Close.
// A second Close is treated as a rewind of the destination so a
// caller that defers Close after an explicit Close empties the
// written σy text.
type ylSink struct {
	dst    io.Writer
	buf    bytes.Buffer
	nclose int
}

func (s *ylSink) Write(p []byte) (int, error) {
	return s.buf.Write(p)
}

func (s *ylSink) Close() error {
	s.nclose++
	if s.nclose == 1 {
		_, err := s.dst.Write(s.buf.Bytes())
		return err
	}
	if b, ok := s.dst.(*bytes.Buffer); ok {
		b.Reset()
	}
	return nil
}

func snapshotYield(sy float64) (out float64) {
	var buf bytes.Buffer
	sink := &ylSink{dst: &buf}
	defer func() {
		_ = sink.Close()
		out, _ = strconv.ParseFloat(strings.TrimSpace(buf.String()), 64)
	}()
	fmt.Fprintf(sink, "%.17g", sy)
	_ = sink.Close()
	return
}
