package main

import (
	"context"
	"log"
	"net/http"

	"github.com/daniel-dihardja/gentic/pkg/chat"
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/daniel-dihardja/gentic/pkg/steps"
	"github.com/joho/godotenv"
)

// chatResolver returns a flow with a single ChatStep so streaming uses the same
// Resolver → Flow path as other agents.
type chatResolver struct {
	model        string
	systemPrompt string
}

func (r chatResolver) Resolve(_ context.Context, _ *gentic.State) gentic.Flow {
	return gentic.NewFlow(steps.ChatStep{
		Model:        r.model,
		SystemPrompt: r.systemPrompt,
	})
}

func main() {
	godotenv.Load()

	provider := openai.Provider{}

	model := openai.DefaultModel
	systemPrompt := "You are a helpful assistant."

	agent := gentic.Agent{
		Resolver: chatResolver{
			model:        model,
			systemPrompt: systemPrompt,
		},
	}

	cfg := chat.Config{
		Agent:        &agent,
		StreamingLLM: provider,
		Model:        model,
		SystemPrompt: systemPrompt,
		AllowOrigins: []string{"http://localhost:3000"},
	}

	mux := http.NewServeMux()
	mux.Handle("/api/chat", chat.CORS(cfg.AllowOrigins, chat.Handler(cfg)))

	log.Println("chat server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
