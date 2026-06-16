package service

import "github.com/mwildt/progoter/request"

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	Messages []*request.Message
}

// NewChatContext creates a new ChatContext with an initial system message.
func NewChatContext() *ChatContext {
	return &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."},
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
func (cc *ChatContext) ClearMessages() {
	cc.Messages = []*request.Message{
		{Role: "system", Content: "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."},
	}
}