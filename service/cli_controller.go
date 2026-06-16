package service

import (
	"bufio"
	"fmt"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"os"
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
	}
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
