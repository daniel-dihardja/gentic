package server

import (
	"net/http"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/middleware"
)

// Config holds dependencies for the default HTTP router.
type Config struct {
	Agent        gentic.Agent
	StreamingLLM gentic.StreamingLLM
	// Model and SystemPrompt are reserved for future use (e.g. CORS or direct LLM routes).
	Model        string
	SystemPrompt string
	AllowOrigins []string
}

// NewRouter wires POST /invoke and POST /invoke/stream with recovery → logging → request ID.
func NewRouter(cfg Config) http.Handler {
	runner := NewRunner(cfg.Agent, cfg.StreamingLLM)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /invoke", InvokeHandler(runner))
	mux.HandleFunc("POST /invoke/stream", StreamHandler(runner))

	return middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.Logging,
		middleware.RequestID,
	)
}
