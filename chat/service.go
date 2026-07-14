package chat

import (
	"context"
	"errors"
	"github.com/mwildt/progoter/chatapi"
	"github.com/mwildt/progoter/tools"
	"log/slog"
)

type Service struct {
	api         *chatapi.Service
	toolService *tools.Service
}

func NewChatService(toolCaller *tools.Service, api *chatapi.Service) *Service {
	return &Service{
		api:         api,
		toolService: toolCaller,
	}
}

func (cs *Service) sendCompleteRequest(ctx context.Context, messages []*chatapi.Message, handler chatapi.MessageHandler) (*chatapi.Message, error) {

	projected := make([]*chatapi.ChatCompletionMessage, len(messages))

	for i, u := range messages {
		projected[i] = &chatapi.ChatCompletionMessage{
			Role:       u.Role,
			ToolCallId: u.ToolCallId,
			ToolCalls:  u.ToolCalls,
			Content:    u.Content,
		}
	}

	var tools []chatapi.Tool
	for _, t := range cs.toolService.GetTools(ctx) {
		tools = append(tools, chatapi.Tool(t))
	}

	request := chatapi.ChatCompletionRequest{
		Model:    "devstral-medium-latest",
		Stream:   true,
		Messages: projected,
		Tools:    tools,
	}

	return cs.api.Complete(ctx, request, handler)
}

func (cs *Service) Complete(ctx context.Context, chatContext *ChatContext) (*ChatContext, error) {
	return cs.CompleteWithHandler(ctx, chatContext, chatapi.MessageHandlerFunc(func(msg *chatapi.Message) {
		chatContext.addMessage(msg)
		chatContext.Broadcast(msg)
	}))
}

func (cs *Service) CompleteWithHandler(ctx context.Context, chatContext *ChatContext, handler chatapi.MessageHandler) (*ChatContext, error) {

	isFirstIteration := true

	for {
		// wurde der context ggf in der zwischenzeit abgebrochen?
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		slog.Default().Info(">> send compoletion reuqets")
		responseMessage, err := cs.sendCompleteRequest(ctx, chatContext.GetMessages(), handler)
		if err != nil {
			return nil, err
		}

		// Beende die Schleife, wenn keine Tool-Calls mehr anstehen
		if len(responseMessage.ToolCalls) == 0 {
			break
		} else {
			// tool calls ausführen und ggf weiter machen
			for _, toolCall := range responseMessage.ToolCalls {
				callContent, err := cs.callTool(ctx, chatContext, toolCall)
				if err != nil {
					slog.Default().Error("Fehler beim Aufruf eines Tools", "tool", toolCall.Type, "error", err)
				}
				toolMessage := &chatapi.Message{
					Role:       "tool",
					ToolCallId: toolCall.Id,
					Content:    string(callContent),
				}
				chatContext.AddMessage(toolMessage)
				chatContext.Broadcast(toolMessage)
			}
		}

		// Ab dem zweiten Durchlauf prüfen, ob der Kontext komprimiert werden muss
		if !isFirstIteration {
			// Aktuelle Kontextgröße aus der letzten Response-Nachricht lesen
			currentContextSize := responseMessage.Usage.TotalTokens
			maxContextSize := 200000

			// Prüfen, ob die Kontextgröße 70% überschreitet
			if maxContextSize > 0 && currentContextSize > 0 {
				percentageUsed := float64(currentContextSize) / float64(maxContextSize)
				if percentageUsed > 0.7 {
					slog.Default().Info("Kontextgröße überschreitet 70%, Komprimierung wird durchgeführt")
					chatContext.Compcat(cs)
				}
			}
		} else {
			isFirstIteration = false
		}
	}
	slog.Default().Info("CompleteWithHandler <<")

	return chatContext, nil
}

func (cs *Service) callTool(ctx context.Context, chatContext *ChatContext, call chatapi.ToolCallChoice) ([]byte, error) {
	slog.Default().Info("Tool CallFunction", "tool", call.Function.Name, "call_id", call.Id)
	if cs.toolService == nil {
		return nil, errors.New("tool caller is nil")
	} else {
		return cs.toolService.CallFunction(ctx, chatContext.BasePath, call.Function.Name, call.Function.Arguments)
	}
}
