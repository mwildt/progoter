package service

import (
	"bufio"
	"fmt"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"os"
)

type CLIController struct {
	chatService  *ChatService
	chatContext  *ChatContext
	messageChan  chan *request.Message
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
	for msg := range cc.messageChan {
		fmt.Print(msg.Content)
	}
}

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
