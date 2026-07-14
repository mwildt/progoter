package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"log/slog"
	"os"
	"strings"
	"time"
)

type CLIController struct {
	chatService *Service
	chatContext *ChatContext
}

func NewCLIController(chatService *Service) *CLIController {
	return &CLIController{
		chatService: chatService,
		chatContext: NewChatContext("./"),
	}
}

// StartChat startet die Chat-Schleife und verwaltet die Benutzereingabe und -ausgabe.
func (cc *CLIController) StartChat() {
	for {
		input := cc.getUserMessage("Was ist dein Begehr")
		inputContent := strings.TrimSpace(input.Content)

		switch inputContent {
		case "/compact":
			if err := cc.CompactChat(); err != nil {
				slog.Error("Fehler beim Komprimieren des Chatverlaufs", "error", err)
			}
		case "/clear":
			if err := cc.ClearContext(); err != nil {
				slog.Error("Fehler beim Zurücksetzen des Chatverlaufs", "error", err)
			}
		case "/dump":
			if err := cc.DumpContext(); err != nil {
				slog.Error("Fehler beim Schreiben des Chatverlaufs", "error", err)
			}

		default:
			cc.chatContext.AddMessage(input)
			messageChan := make(chan *chatapi.Message)
			go func() {
				var err error
				cc.chatContext, err = cc.chatService.CompleteWithHandler(context.Background(), cc.chatContext, chatapi.MessageHandlerFunc(func(msg *chatapi.Message) {
					messageChan <- msg
				}))
				if err != nil {
					slog.Error("Fehler beim Verarbeiten der Chat-Vervollständigung", "error", err)
				}
			}()
			cc.listenForMessages(messageChan)
		}
	}
}

// CompactChat komprimiert den aktuellen Chatverlauf, indem eine Zusammenfassung vom LLM angefordert wird.
func (cc *CLIController) CompactChat() error {
	// Create a new message to request summarization
	summarizeMessage := &chatapi.Message{
		Role: "user",
		Content: `
Fasse den bisherigen Chatverlauf zusammen. Ziel ist es, **alle fachlichen Informationen, Entscheidungen, Daten, Code-Snippets und Kontext** zu erhalten. Ignoriere dabei:
- Smalltalk
- Bestätigungen ("Ja", "Okay", "Verstanden")
- Wiederholungen
- Off-Topic-Diskussionen

**Regeln:**
0. Ignoriere System nachrichten .
1. Behalte **technische Details, Anforderungen, Lösungsansätze und offene Fragen** bei.
2. Strukturiere die Zusammenfassung nach Themen/Abschnitten (z. B. "Anforderungen", "Technische Umsetzung", "Offene Punkte").
3. Verwende **die gleiche Terminologie** wie im Original.
4. Falls Code oder Daten im Verlauf vorkommen: **Füge sie unverändert ein** (keine Paraphrasierung).
5. Maximal 50 % der ursprünglichen Token-Anzahl.

**Ausgabeformat:**
- Knappe, klare Sätze.
- Keine Erklärungen, warum etwas zusammengefasst wurde.
- Keine Einleitungen wie "Hier ist die Zusammenfassung:".
`,
	}

	// Add the summarize message to the current context
	cc.chatContext.AddMessage(summarizeMessage)

	// No need for a message channel here as we don't stream messages to the user
	var messageChan chan *chatapi.Message = nil

	// Request completion from the chat service
	compactedContext, err := cc.chatService.CompleteWithHandler(context.Background(), cc.chatContext, chatapi.MessageHandlerFunc(func(msg *chatapi.Message) {
		messageChan <- msg
	}))
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
		Messages: []*chatapi.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "assistant", Content: summary},
		},
	}

	cc.chatContext = newContext
	fmt.Println("\nChatverlauf wurde erfolgreich komprimiert.")
	return nil
}

// DumpContext speichert den aktuellen Chat-Context als JSON in einer Datei.
func (cc *CLIController) DumpContext() error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("dumps/context.%s.json", timestamp)

	// Erstelle das Verzeichnis, falls es nicht existiert
	err := os.MkdirAll("dumps", os.ModePerm)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen des Verzeichnisses: %v", err)
	}

	// Öffne die Datei zum Schreiben
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen der Datei: %v", err)
	}
	defer file.Close()

	// Konvertiere den Chat-Context in JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cc.chatContext)
	if err != nil {
		return fmt.Errorf("Fehler beim Schreiben des JSON: %v", err)
	}

	fmt.Printf("\nChat-Context wurde erfolgreich in %s gespeichert.\n", filename)
	return nil
}

// ClearContext setzt den Chat-Context zurück und erstellt einen neuen Kontext mit dem Standard-Prompt.
func (cc *CLIController) ClearContext() error {
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	// Create a new context with the system prompt
	newContext := &ChatContext{
		Messages: []*chatapi.Message{
			{Role: "system", Content: systemPrompt},
		},
	}

	cc.chatContext = newContext
	fmt.Println("\nChat-Context wurde erfolgreich zurückgesetzt.")
	return nil
}

// listenForMessages hört auf Nachrichten im Channel und gibt sie aus.
func (cc *CLIController) listenForMessages(messageChan chan *chatapi.Message) {
	var lastRole string

	for msg := range messageChan {
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
func (cc *CLIController) getUserMessage(s string) *chatapi.Message {
	reader := bufio.NewReader(os.Stdin)
	println(s)
	input, _ := reader.ReadString('\n')
	return &chatapi.Message{
		Role:    "user",
		Content: input,
	}
}
