package main

import (
	"log"
	"net/http"

	"github.com/daniel-dihardja/gentic/pkg/chat"
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/joho/godotenv"
)

// directResolver routes every input to a single LLM step via the streaming path.
// The Flow/Resolver pipeline is bypassed by StreamWithContext, so this resolver
// is only needed to satisfy the Agent struct — it is never called during streaming.
type directResolver struct{}

func (directResolver) Resolve(_ *gentic.State) gentic.Flow {
	return gentic.NewFlow()
}

func main() {
	godotenv.Load()

	provider := openai.Provider{}

	agent := gentic.Agent{
		Resolver: directResolver{},
	}

	cfg := chat.Config{
		Agent:        &agent,
		StreamingLLM: provider,
		Model:        openai.DefaultModel,
		SystemPrompt: "You are a helpful assistant.",
		AllowOrigins: []string{"http://localhost:3000"},
	}

	mux := http.NewServeMux()
	mux.Handle("/api/chat", chat.CORS(cfg.AllowOrigins, chat.Handler(cfg)))

	log.Println("chat server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
