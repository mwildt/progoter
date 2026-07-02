package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/mwildt/progoter/request"
)

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	BasePath    string
	Messages    []*request.Message `json:"messages"`
	IsStreaming bool               `json:"is_streaming"`
	State       string             `json:"state"`
	subscribers map[chan *request.Message]bool
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
	slog.Default().Info("NewChatContext")
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
		IsStreaming: false,
		State:       "idle",
		subscribers: make(map[chan *request.Message]bool),
	}
}

// AddMessage adds a message to the chat context.
func (cc *ChatContext) AddMessage(message *request.Message) {
	slog.Default().Info("ChatContext::AddMessage")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.addMessage(message)
}

func (cc *ChatContext) Complete(service *ChatService) error {
	slog.Default().Info("ChatContext::Complete")
	cc.mu.Lock()
	cc.State = "running"
	cc.mu.Unlock()
	_, err := service.Complete(context.Background(), cc)
	cc.mu.Lock()
	cc.State = "idle"
	cc.mu.Unlock()
	return err
}

func (cc *ChatContext) addMessage(message *request.Message) {
	if len(cc.Messages) == 0 {
		cc.Messages = append(cc.Messages, message)
	} else if last := cc.Messages[len(cc.Messages)-1]; message.Role == last.Role {
		last.Content += message.Content
		for _, tc := range message.ToolCalls {
			last.ToolCalls = append(last.ToolCalls, tc)
		}
		last.Usage = message.Usage
	} else {
		cc.Messages = append(cc.Messages, message)
	}
}

// AddMessages adds multiple messages to the chat context.
func (cc *ChatContext) AddMessages(messages []*request.Message) {
	slog.Default().Info("ChatContext::AddMessages")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for _, message := range messages {
		cc.addMessage(message)
	}
}

func (cc *ChatContext) Stream() chan *request.Message {
	slog.Default().Info("ChatContext::Stream")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	sub := make(chan *request.Message, len(cc.Messages))
	for _, msg := range cc.Messages {
		sub <- msg
	}
	cc.subscribers[sub] = true
	cc.BroadcastState()
	return sub
}

// Unsubscribe entfernt einen Abonnenten.
func (cc *ChatContext) Unsubscribe(sub chan *request.Message) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	close(sub)
	delete(cc.subscribers, sub)
}

func (cc *ChatContext) CloseSubscriptions() {
	slog.Default().Info("CloseSubscriptions")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for sub := range cc.subscribers {
		delete(cc.subscribers, sub)
		close(sub)
	}
}

// Broadcast sendet eine Nachricht an alle Abonnenten.
func (cc *ChatContext) Broadcast(msg *request.Message) {
	slog.Default().Info("Broadcast")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for sub := range cc.subscribers {
		sub <- msg
	}
}

func (cc *ChatContext) BroadcastState() {
	slog.Default().Info("BroadcastState")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for sub := range cc.subscribers {
		sub <- &request.Message{
			Role:    "system",
			Content: fmt.Sprintf(`{"state": "%s"}`, cc.State),
		}
	}
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
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.IsStreaming = false
	cc.State = "idle"
	for sub := range cc.subscribers {
		close(sub)
		delete(cc.subscribers, sub)
	}
	return nil
}
