package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// sySink buffers a Hall-Petch /api/sy encode and flushes on Close.
// A second Close is treated as a rewind of the destination so a
// caller that defers Close after an explicit Close empties the
// written JSON body.
type sySink struct {
	dst    io.Writer
	buf    bytes.Buffer
	nclose int
}

func (s *sySink) Write(p []byte) (int, error) {
	return s.buf.Write(p)
}

func (s *sySink) Close() error {
	s.nclose++
	if s.nclose == 1 {
		_, err := s.dst.Write(s.buf.Bytes())
		return err
	}
	return nil
}

// writeSyBody encodes the yield-strength response through sySink.
func writeSyBody(w http.ResponseWriter, resp syResponse) {
	var buf bytes.Buffer
	sink := &sySink{dst: &buf}
	defer func() {
		_ = sink.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}()
	_ = json.NewEncoder(sink).Encode(resp)
	_ = sink.Close()
}
