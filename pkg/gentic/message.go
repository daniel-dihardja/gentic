package gentic

import (
	"strings"
	"time"
)

// MessagePart represents a single content part within a message.
// Compatible with Vercel AI SDK UIMessage format.
type MessagePart struct {
	Type   string      `json:"type"`             // "text", "tool-*"
	Text   string      `json:"text,omitempty"`   // text content
	Input  interface{} `json:"input,omitempty"`  // tool input
	Output interface{} `json:"output,omitempty"` // tool output
}

// Message represents a single message in a conversation.
// Compatible with Vercel AI SDK UIMessage format.
type Message struct {
	ID        string        `json:"id,omitempty"`
	Role      string        `json:"role"` // "user", "assistant", "system"
	Parts     []MessagePart `json:"parts"`
	CreatedAt time.Time     `json:"createdAt,omitempty"`
}

// TextContent extracts the plain text content from this message.
// Returns concatenated text from all text-type parts.
func (m Message) TextContent() string {
	var sb strings.Builder
	for i, part := range m.Parts {
		if part.Type == "text" && part.Text != "" {
			if i > 0 && sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(part.Text)
		}
	}
	return sb.String()
}

// NewUserMessage creates a user message with plain text content.
func NewUserMessage(text string) Message {
	return Message{
		ID:        generateID(),
		Role:      "user",
		Parts:     []MessagePart{{Type: "text", Text: text}},
		CreatedAt: time.Now(),
	}
}

// NewAssistantMessage creates an assistant message with plain text content.
func NewAssistantMessage(text string) Message {
	return Message{
		ID:        generateID(),
		Role:      "assistant",
		Parts:     []MessagePart{{Type: "text", Text: text}},
		CreatedAt: time.Now(),
	}
}

// NewSystemMessage creates a system message with plain text content.
func NewSystemMessage(text string) Message {
	return Message{
		ID:        generateID(),
		Role:      "system",
		Parts:     []MessagePart{{Type: "text", Text: text}},
		CreatedAt: time.Now(),
	}
}

// generateID creates a simple message ID.
func generateID() string {
	return "msg-" + time.Now().Format("20060102150405") + "-" + randStr(8)
}

// randStr generates a random string of n characters.
func randStr(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[int64(i)%int64(len(charset))]
	}
	return string(b)
}
