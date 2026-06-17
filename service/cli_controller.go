package service

import (
	"bufio"
	"fmt"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type CLIController struct {
	chatService *ChatService
	chatContext *ChatContext
	messageChan chan *request.Message
}

func NewCLIController(apiKey string) *CLIController {
	return &CLIController{
		chatService: NewChatService(apiKey),
		chatContext: NewChatContext(),
		messageChan: make(chan *request.Message),
	}
}

// StartChat startet die Chat-Schleife und verwaltet die Benutzereingabe und -ausgabe.
func (cc *CLIController) StartChat() {
	go cc.listenForMessages()
	for {
		cc.chatContext.AddMessage(cc.getUserMessage("Was ist dein Begehr"))

		var err error
		cc.chatContext, err = cc.chatService.CompleteContext(cc.chatContext, cc.messageChan)
		if err != nil {
			slog.Error("Fehler beim Verarbeiten der Chat-Vervollständigung", "error", err)
		}
		fmt.Printf("TotalTokens: %d\n", cc.chatContext.TotalTokens)
	}
}

// CompactChat komprimiert den aktuellen Chatverlauf, indem eine Zusammenfassung vom LLM angefordert wird.
func (cc *CLIController) CompactChat() error {
	// Create a new message to request summarization
	summarizeMessage := &request.Message{
		Role:    "user",
		Content: "Fasse den bisherigen Chatverlauf zusammen, um den Kontext zu komprimieren. Halte dabei die wichtigsten Informationen fest.",
	}

	// Add the summarize message to the current context
	cc.chatContext.AddMessage(summarizeMessage)

	// Create a channel for messages (not used here, but required by CompleteContext)
	messageChan := make(chan *request.Message)

	// Request completion from the chat service
	compactedContext, err := cc.chatService.CompleteContext(cc.chatContext, messageChan)
	if err != nil {
		return fmt.Errorf("Fehler beim Komprimieren des Chatverlaufs: %v", err)
	}

	// Extract the summary from the last assistant message
	var summary string
	messages := compactedContext.GetMessages()
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			summary = messages[i].Content
			break
		}
	}

	// Create a new context with the system prompt and the summary
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	newContext := &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "assistant", Content: summary},
		},
	}

	cc.chatContext = newContext
	fmt.Println("\nChatverlauf wurde erfolgreich komprimiert.")
	return nil
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

// listenForMessages hört auf Nachrichten im Channel und gibt sie aus.
func (cc *CLIController) listenForMessages() {
	var lastRole string

	for msg := range cc.messageChan {
		if len(strings.TrimSpace(msg.Content)) == 0 {
			continue
		}
		if msg.Role != lastRole {
			color := getColorForRole(msg.Role)
			fmt.Printf("\n%s##########################\n[%s]\n%s", color, msg.Role, resetColor)
			lastRole = msg.Role
		}
		color := getColorForRole(msg.Role)
		fmt.Printf("%s%s%s", color, msg.Content, resetColor)
	}
}

// getColorForRole gibt den ANSI-Farbcode für die gegebene Rolle zurück.
func getColorForRole(role string) string {
	switch role {
	case "user":
		return "\033[32m" // Grün für Benutzer
	case "assistant":
		return "\033[34m" // Blau für Assistent
	case "tool":
		return "\033[38;5;208m" // Orange für Tool
	default:
		return "\033[0m" // Standardfarbe für unbekannte Rollen
	}
}

const resetColor = "\033[0m"

// getUserMessage liest eine Nachricht vom Benutzer.
func (cc *CLIController) getUserMessage(s string) *request.Message {
	reader := bufio.NewReader(os.Stdin)
	println(s)
	input, _ := reader.ReadString('\n')
	return &request.Message{
		Role:    "user",
		Content: input,
	}
}
