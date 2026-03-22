package gentic

import (
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test append
	msg1 := NewUserMessage("Hello")
	msg2 := NewAssistantMessage("Hi there!")
	msg3 := NewUserMessage("How are you?")

	if err := storage.Append(msg1); err != nil {
		t.Fatalf("failed to append msg1: %v", err)
	}
	if err := storage.Append(msg2); err != nil {
		t.Fatalf("failed to append msg2: %v", err)
	}
	if err := storage.Append(msg3); err != nil {
		t.Fatalf("failed to append msg3: %v", err)
	}

	// Test messages retrieval
	messages, err := storage.Messages()
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	// Verify order
	if messages[0].TextContent() != "Hello" {
		t.Errorf("expected first message to be 'Hello', got '%s'", messages[0].TextContent())
	}
	if messages[1].TextContent() != "Hi there!" {
		t.Errorf("expected second message to be 'Hi there!', got '%s'", messages[1].TextContent())
	}
	if messages[2].TextContent() != "How are you?" {
		t.Errorf("expected third message to be 'How are you?', got '%s'", messages[2].TextContent())
	}

	// Test clear
	if err := storage.Clear(); err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	messages, err = storage.Messages()
	if err != nil {
		t.Fatalf("failed to get messages after clear: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(messages))
	}
}

func TestMessageTextContent(t *testing.T) {
	msg := Message{
		Role: "user",
		Parts: []MessagePart{
			{Type: "text", Text: "Hello"},
			{Type: "text", Text: "world"},
		},
	}

	content := msg.TextContent()
	if content != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", content)
	}
}

func TestNewUserMessage(t *testing.T) {
	msg := NewUserMessage("Test query")
	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}
	if msg.TextContent() != "Test query" {
		t.Errorf("expected text 'Test query', got '%s'", msg.TextContent())
	}
	if msg.ID == "" {
		t.Error("expected non-empty ID")
	}
}
