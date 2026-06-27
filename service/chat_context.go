package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mwildt/progoter/request"
)

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	BasePath    string
	Messages    []*request.Message `json:"messages"`
	TotalTokens int                `json:"total_tokens"`
	messageChan chan *request.Message
	sseClients  map[chan string]bool
	mu          sync.Mutex
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
	chatContext := &ChatContext{
		BasePath:    "./",
		Messages:    []*request.Message{
			{Role: "system", Content: systemPrompt},
		},
		messageChan: make(chan *request.Message),
		sseClients:  make(map[chan string]bool),
	}
	chatContext.StartBroadcasting()
	return chatContext
}

// AddMessage adds a message to the chat context and sends it to all active SSE clients.
func (cc *ChatContext) AddMessage(message *request.Message) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Messages = append(cc.Messages, message)
	cc.messageChan <- message
}

// AddMessages adds multiple messages to the chat context.
func (cc *ChatContext) AddMessages(messages []*request.Message) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Messages = append(cc.Messages, messages...)
}

// GetMessages returns all messages in the chat context.
func (cc *ChatContext) GetMessages() []*request.Message {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.Messages
}

// ClearMessages clears all messages in the chat context.
func (cc *ChatContext) ClearMessages() error {
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		// Fallback to default system prompt if file reading fails
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Messages = []*request.Message{
		{Role: "system", Content: systemPrompt},
	}
	return nil
}

// RegisterSSEClient registriert einen neuen SSE-Client.
func (cc *ChatContext) RegisterSSEClient(clientChan chan string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.sseClients[clientChan] = true
}

// UnregisterSSEClient entfernt einen SSE-Client.
func (cc *ChatContext) UnregisterSSEClient(clientChan chan string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.sseClients, clientChan)
	close(clientChan)
}

// Broadcast sendet eine Nachricht an alle aktiven SSE-Clients.
func (cc *ChatContext) Broadcast(message *request.Message) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	data, _ := json.Marshal(message)
	event := fmt.Sprintf("data: %s\n\n", string(data))
	for clientChan := range cc.sseClients {
		clientChan <- event
	}
}

// StartBroadcasting startet das Broadcasting von Nachrichten an SSE-Clients.
func (cc *ChatContext) StartBroadcasting() {
	go func() {
		for msg := range cc.messageChan {
			cc.Broadcast(msg)
		}
	}()
}
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}
	cc.Messages = []*request.Message{
		{Role: "system", Content: systemPrompt},
	}
	return nil
}
