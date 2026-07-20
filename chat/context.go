package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"log/slog"
	"os"
	"sync"
	"time"
)

type StateEvent int

const (
	StateProcessing StateEvent = iota
	StateIdle
)

type PubSub[T any] struct {
	subscribers map[chan T]bool
	mu          sync.Mutex
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		subscribers: make(map[chan T]bool),
	}
}

func (p *PubSub[T]) Subscribe(bufferSize int) chan T {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub := make(chan T, bufferSize)
	p.subscribers[sub] = true
	return sub
}

func (p *PubSub[T]) Unsubscribe(ch chan T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(ch)
	delete(p.subscribers, ch)
}

func (p *PubSub[T]) CloseSubscriptions() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

func (p *PubSub[T]) Broadcast(msg T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for sub := range p.subscribers {
		sub <- msg
	}
}

// ChatContext represents a collection of messages in a chat.
type ChatContext struct {
	BasePath string             // das ist ein relativer Pfad
	Messages []*chatapi.Message `json:"messages"`
	cancel   context.CancelFunc
	pubSub   *PubSub[any]
	mu       sync.Mutex
}

// readSystemPrompt liest den System-Prompt aus einer Datei.
func readSystemPrompt() (string, error) {
	data, err := os.ReadFile("prompts/system-default.md")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// readSystemPrompt liest den System-Prompt aus einer Datei.
func readCompcatDefaultPrompt() (string, error) {
	data, err := os.ReadFile("prompts/default-compaction.md")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewChatContext creates a new ChatContext with an initial system message.
func NewChatContext(basePath string) *ChatContext {
	slog.Default().Info("NewChatContext")
	// Read the system prompt from the file
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		// Fallback to default system prompt if file reading fails
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	return &ChatContext{
		BasePath: basePath,
		Messages: []*chatapi.Message{
			{Role: "system", Content: systemPrompt},
		},
		pubSub: NewPubSub[any](),
	}
}

// AddMessage adds a message to the chat context.
func (cc *ChatContext) AddMessage(message *chatapi.Message) {
	slog.Default().Info("ChatContext::AddMessage", "role", message.Role, "content", message.Content)
	//cc.mu.Lock()
	//defer cc.mu.Unlock()
	cc.Messages = append(cc.Messages, message)
}

func (cc *ChatContext) Complete(service *Service) error {
	slog.Default().Info("ChatContext::Complete")
	cc.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	cc.cancel = cancel
	defer func() {
		cc.Broadcast(StateIdle)
		cc.cancel = nil
		cc.mu.Unlock()
	}()
	cc.Broadcast(StateProcessing)
	_, err := service.Complete(ctx, cc)
	return err
}

func (cc *ChatContext) Cancel() {
	if cc.cancel == nil {
		return
	}
	slog.Default().Info("ChatContext::Cancel")
	cc.cancel()
	cc.cancel = nil
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for i := len(cc.Messages) - 1; i >= 0; i-- {
		lastMessage := cc.Messages[i]
		if lastMessage.HasRole("user") {
			cc.Messages = cc.Messages[:i]
			return
		}
	}
}

func (cc *ChatContext) addMessage(message *chatapi.Message) {
	if len(cc.Messages) == 0 {
		cc.Messages = append(cc.Messages, chatapi.FromMessage(message))
	} else if last := cc.Messages[len(cc.Messages)-1]; last.HasRole(message.Role) {
		last.Join(message)
	} else {
		cc.Messages = append(cc.Messages, chatapi.FromMessage(message))
	}
}

// AddMessages adds multiple messages to the chat context.
func (cc *ChatContext) AddMessages(messages []*chatapi.Message) {
	slog.Default().Info("ChatContext::AddMessages")
	cc.mu.Lock()
	defer cc.mu.Unlock()
	for _, message := range messages {
		cc.addMessage(message)
	}
}

func (cc *ChatContext) Stream() chan any {
	slog.Default().Info("ChatContext::Stream")
	sub := cc.pubSub.Subscribe(len(cc.Messages) + 1)
	if cc.cancel != nil {
		sub <- StateProcessing
	}
	for _, msg := range cc.Messages {
		sub <- msg
	}
	return sub
}

// Unsubscribe entfernt einen Abonnenten.
func (cc *ChatContext) Unsubscribe(sub chan any) {
	cc.pubSub.Unsubscribe(sub)
}

func (cc *ChatContext) CloseSubscriptions() {
	slog.Default().Info("CloseSubscriptions")
	cc.pubSub.CloseSubscriptions()
}

// Broadcast sendet eine Nachricht an alle Abonnenten.
func (cc *ChatContext) Broadcast(msg any) {
	cc.pubSub.Broadcast(msg)
}

// GetMessages returns all messages in the chat context.
func (cc *ChatContext) GetMessages() []*chatapi.Message {
	return cc.Messages
}

func (cc *ChatContext) IsStoppend() bool {
	return nil == cc.cancel
}

func (cc *ChatContext) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Messages = cc.Messages[:1]
}

func (cc *ChatContext) Dump() error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("dumps/context.%s.json", timestamp)

	err := os.MkdirAll("dumps", os.ModePerm)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen des Verzeichnisses: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen der Datei: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cc)
	if err != nil {
		return fmt.Errorf("Fehler beim Schreiben des JSON: %v", err)
	}
	return nil
}

func (cc *ChatContext) Compcat(service *Service) error {

	prompt, err := readCompcatDefaultPrompt()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cc.mu.Lock()
	defer func() {
		cc.Broadcast(StateIdle)
		cc.mu.Unlock()
		cc.cancel = nil
	}()
	cc.cancel = cancel
	cc.Broadcast(StateProcessing)

	maxCompactSize := 200000
	compactThreshold := int(0.4 * float64(maxCompactSize))

	// Create a new compaction context with messages 1:n
	compactionContext := &ChatContext{
		BasePath: cc.BasePath,
		Messages: []*chatapi.Message{},
		pubSub:   NewPubSub[any](),
	}

	lastTotalUsage := 0
	messageRangeMange := 0
	for _, msg := range cc.Messages {
		// das hier geht nur, weil wir wissen, dass nur bei assistent messages eine Usage gesetzt ist.
		if msg.Usage.TotalTokens > 0 {
			lastTotalUsage += msg.Usage.TotalTokens
		}
		if lastTotalUsage < compactThreshold {
			compactionContext.AddMessage(chatapi.FromMessage(msg))
			messageRangeMange += 1
		} else {
			break
		}
	}

	compactionContext.addMessage(&chatapi.Message{
		Role:    "user",
		Content: prompt,
	})

	// Perform compaction on the new context
	summary := &chatapi.Message{
		Role: "user",
	}
	_, err = service.CompleteWithHandler(ctx, compactionContext, summary)
	if err != nil {
		return fmt.Errorf("Fehler beim Kompaktieren: %v", err)
	}

	newMessages := append([]*chatapi.Message{}, cc.Messages[1])           // alter system promt
	newMessages = append(newMessages, summary)                            // zusammenfassung
	newMessages = append(newMessages, cc.Messages[messageRangeMange:]...) // weitere nachrichten

	cc.Messages = newMessages
	return nil
}
