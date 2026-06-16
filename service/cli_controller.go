package service

import (
	"bufio"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"os"
)

type CLIController struct {
	chatService *ChatService
	chatContext *ChatContext
}

func NewCLIController(apiKey string) *CLIController {
	return &CLIController{
		chatService: NewChatService(apiKey),
		chatContext: NewChatContext(),
	}
}

// StartChat startet die Chat-Schleife und verwaltet die Benutzereingabe.
func (cc *CLIController) StartChat() {
	for {
		cc.chatContext.AddMessage(cc.getUserMessage("Was ist dein Begehr"))

		var err error
		cc.chatContext, err = cc.chatService.CompleteContext(cc.chatContext)
		if err != nil {
			slog.Error("Fehler beim Verarbeiten der Chat-Vervollständigung", "error", err)
		}
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
