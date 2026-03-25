// Package sse provides helpers for writing Server-Sent Events over HTTP.
package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// ErrNoFlusher is returned when the ResponseWriter does not implement [http.Flusher].
var ErrNoFlusher = errors.New("sse: response writer does not support flushing")

// Writer wraps an [http.ResponseWriter] for SSE frames (data: …\n\n) with flush.
type Writer struct {
	w http.ResponseWriter
	f http.Flusher
}

// New sets standard SSE response headers and returns a Writer.
// Returns [ErrNoFlusher] if w does not implement [http.Flusher].
func New(w http.ResponseWriter) (*Writer, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrNoFlusher
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	return &Writer{w: w, f: f}, nil
}

// WriteData writes one SSE frame: data: <payload>\n\n and flushes.
func (sw *Writer) WriteData(payload []byte) {
	fmt.Fprintf(sw.w, "data: %s\n\n", payload)
	sw.f.Flush()
}

// WriteDone writes the sentinel frame data: [DONE]\n\n and flushes.
func (sw *Writer) WriteDone() {
	fmt.Fprintf(sw.w, "data: [DONE]\n\n")
	sw.f.Flush()
}

// Drain consumes ch until a Done token, a stream error, channel close, or ctx cancellation.
// Text tokens are sent as JSON {"delta":"..."}; errors as {"error":"..."}.
// Returns [context.Canceled] / [context.DeadlineExceeded] when ctx ends first;
// the token error when the stream reports an error; nil on clean completion ([Done] or empty close).
func (sw *Writer) Drain(ctx context.Context, ch <-chan gentic.StreamEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-ch:
			if !ok {
				sw.WriteDone()
				return nil
			}
			if ev.Token.Error != nil {
				b, err := json.Marshal(map[string]string{"error": ev.Token.Error.Error()})
				if err != nil {
					return err
				}
				sw.WriteData(b)
				return ev.Token.Error
			}
			if ev.Activity != nil {
				b, err := json.Marshal(map[string]interface{}{"activity": ev.Activity})
				if err != nil {
					return err
				}
				sw.WriteData(b)
			}
			if ev.Token.Text != "" {
				b, err := json.Marshal(map[string]string{"delta": ev.Token.Text})
				if err != nil {
					return err
				}
				sw.WriteData(b)
			}
			if ev.Token.Done {
				sw.WriteDone()
				return nil
			}
		}
	}
}
