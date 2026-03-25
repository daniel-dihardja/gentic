package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/daniel-dihardja/gentic/pkg/sse"
)

// InvokeHandler handles POST /invoke — validates JSON, runs the agent, returns JSON.
func InvokeHandler(runner *Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		var req InvokeRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if req.Message == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}

		resp, err := runner.Invoke(r.Context(), req)
		if err != nil {
			log.Printf("invoke: %v", err)
			writeError(w, http.StatusInternalServerError, "agent run failed")
			return
		}

		writeJSON(w, http.StatusOK, *resp)
	}
}

// StreamHandler handles POST /invoke/stream — SSE with {"delta":"..."} lines and data: [DONE].
func StreamHandler(runner *Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		var req InvokeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if req.Message == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}

		sw, err := sse.New(w)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		events, err := runner.Stream(r.Context(), req)
		if err != nil {
			log.Printf("stream: %v", err)
			b, mErr := json.Marshal(map[string]string{"error": err.Error()})
			if mErr != nil {
				log.Printf("sse marshal: %v", mErr)
				return
			}
			sw.WriteData(b)
			return
		}

		if err := sw.Drain(r.Context(), events); err != nil {
			log.Printf("stream drain: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, InvokeResponse{
		OK:    false,
		Error: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, v InvokeResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}
