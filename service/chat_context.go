package service

import (
	"github.com/mwildt/progoter/request"
	"os"
	"path/filepath"
	"strings"
)

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	Messages []*request.Message
}

// NewChatContext creates a new ChatContext with an initial system message.
func NewChatContext() *ChatContext {
	// Read the system prompt from the file
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		// Fallback to default system prompt if file reading fails
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}
	return &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: systemPrompt},
		},
	}
}

// AddMessage adds a message to the chat context.
func (cc *ChatContext) AddMessage(message *request.Message) {
	cc.Messages = append(cc.Messages, message)
}

// AddMessages adds multiple messages to the chat context.
func (cc *ChatContext) AddMessages(messages []*request.Message) {
	cc.Messages = append(cc.Messages, messages...)
}

// readSystemPrompt reads the system prompt from the prompts/system-default.md file.
func readSystemPrompt() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Construct the path to the system prompt file
	promptPath := filepath.Join(cwd, "prompts", "system-default.md")

	// Read the file content
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}

	// Convert the content to a string and trim any leading/trailing whitespace
	return strings.TrimSpace(string(content)), nil
}

// GetMessages returns all messages in the chat context.
func (cc *ChatContext) GetMessages() []*request.Message {
	return cc.Messages
}

// ClearMessages clears all messages in the chat context.
func (cc *ChatContext) ClearMessages() {
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		// Fallback to default system prompt if file reading fails
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}
	cc.Messages = []*request.Message{
		{Role: "system", Content: systemPrompt},
	}
}