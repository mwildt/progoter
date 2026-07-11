package chat

import (
	"testing"
)

func TestClearContext(t *testing.T) {
	// Create a new CLIController
	cc := NewCLIController("test-api-key")

	// Add some messages to the context
	cc.chatContext.AddMessage(&Message{
		Role:    "user",
		Content: "Hello, world!",
	})
	cc.chatContext.AddMessage(&Message{
		Role:    "assistant",
		Content: "Hi there!",
	})

	// Verify that the context has messages
	if len(cc.chatContext.Messages) != 3 { // 1 system + 2 user/assistant
		t.Errorf("Expected 3 messages before clear, got %d", len(cc.chatContext.Messages))
	}

	// Clear the context
	err := cc.ClearContext()
	if err != nil {
		t.Errorf("Error clearing context: %v", err)
	}

	// Verify that the context has only the system message
	if len(cc.chatContext.Messages) != 1 {
		t.Errorf("Expected 1 message after clear, got %d", len(cc.chatContext.Messages))
	}

	// Verify that the remaining message is a system message
	if cc.chatContext.Messages[0].Role != "system" {
		t.Errorf("Expected first message to be a system message, got %s", cc.chatContext.Messages[0].Role)
	}

	// Verify that the system message is not empty
	if len(cc.chatContext.Messages[0].Content) == 0 {
		t.Errorf("Expected system message to have content")
	}
}
