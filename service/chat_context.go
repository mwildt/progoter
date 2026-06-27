package service

import (
	"os"

	"github.com/mwildt/progoter/request"
)

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	BasePath    string
	Messages    []*request.Message `json:"messages"`
	TotalTokens int                `json:"total_tokens"`
}

// readSystemPrompt liest den System-Prompt aus einer Datei.
func readSystemPrompt() (string, error) {
	data, err := os.ReadFile("system_prompt.txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
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
		BasePath: "./",
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

// GetMessages returns all messages in the chat context.
func (cc *ChatContext) GetMessages() []*request.Message {
	return cc.Messages
}

// ClearMessages clears all messages in the chat context.
func (cc *ChatContext) ClearMessages() error {
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		// Fallback to default system prompt if file reading fails
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}
	cc.Messages = []*request.Message{
		{Role: "system", Content: systemPrompt},
	}
	return nil
}
