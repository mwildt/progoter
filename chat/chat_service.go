package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/mwildt/progoter/chatapi"

	// Füge die letzten 5 Nachrichten wieder hinzu
	"fmt"
	"github.com/mwildt/progoter/tools"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ChatService struct {
	apiKey      string
	client      *http.Client
	toolService *tools.Service
}

func NewChatService(apiKey string, toolCaller *tools.Service) *ChatService {
	return &ChatService{
		apiKey:      apiKey,
		client:      &http.Client{},
		toolService: toolCaller,
	}
}

func readResponse(body io.Reader, handler MessageHandler) (result *Message, err error) {
	reader := bufio.NewReader(body)
	result = &Message{}
	var builder strings.Builder
	var completition chatapi.CompletionChunk

	for {
		// lesen des Response
		var line string
		line, err = reader.ReadString('\n')
		if err != nil {
			return result, err
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}

			err = json.Unmarshal([]byte(data), &completition)
			if err != nil {
				return result, err
			}
			choice := completition.Choices[0]

			result.Role = choice.Delta.Role

			var contentPart string
			// content
			switch content := choice.Delta.Content.(type) {
			case map[string]any:
				fmt.Printf("MAP %v\n", content)
			case []any:
				fmt.Printf("LIST %v\n", content)
			case string:
				contentPart = content
				builder.WriteString(content)
			}

			result.Usage = chatapi.Usage(completition.Usage)

			// tool-calls
			if len(choice.Delta.ToolCalls) > 0 {
				result.ToolCalls = append(result.ToolCalls, choice.Delta.ToolCalls...)
			}
			if nil != handler {
				handler.Join(&Message{
					Role:      choice.Delta.Role,
					ToolCalls: choice.Delta.ToolCalls,
					Content:   contentPart,
					Usage:     chatapi.Usage(completition.Usage),
				})
			}
		}
	}

	slog.Default().Info("EOD")

	result.Content = builder.String()
	return result, nil
}

func (cs *ChatService) sendCompleteRequest(ctx context.Context, messages []*Message, handler MessageHandler) (*Message, error) {

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

	jsonData, err := json.Marshal(&chatapi.ChatCompletionRequest{
		Model:    "devstral-medium-latest",
		Stream:   true,
		Messages: projected,
		Tools:    tools,
	})

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	slog.Default().With("logger", "ChatService").
		Info("send completion request", "url", req.URL.String())

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Fehler beim Lesen des Response-Body", "error", err)
			resp.Body.Close()
			return nil, err
		}

		resp.Body = io.NopCloser(bytes.NewBuffer(body))

		slog.Error("HTTP-Fehler", "status", resp.StatusCode, "body", string(body))
		resp.Body.Close()
		return nil, errors.New("HTTP-Fehler")
	}

	return readResponse(resp.Body, handler)
}

type MessageHandler interface {
	Join(*Message)
}

type MessageHandlerFunc func(*Message)

func (fn MessageHandlerFunc) Join(message *Message) {
	fn(message)
}

func (cs *ChatService) Complete(ctx context.Context, chatContext *ChatContext) (*ChatContext, error) {
	return cs.CompleteWithHandler(ctx, chatContext, MessageHandlerFunc(func(msg *Message) {
		chatContext.addMessage(msg)
		chatContext.Broadcast(msg)
	}))
}

func (cs *ChatService) CompleteWithHandler(ctx context.Context, chatContext *ChatContext, handler MessageHandler) (*ChatContext, error) {

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
				toolMessage := &Message{
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

func (cs *ChatService) callTool(ctx context.Context, chatContext *ChatContext, call chatapi.ToolCallChoice) ([]byte, error) {
	slog.Default().Info("Tool CallFunction", "tool", call.Function.Name, "call_id", call.Id)
	if cs.toolService == nil {
		return nil, errors.New("tool caller is nil")
	} else {
		return cs.toolService.CallFunction(ctx, chatContext.BasePath, call.Function.Name, call.Function.Arguments)
	}
}
