package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	// Füge die letzten 5 Nachrichten wieder hinzu
	"fmt"
	"github.com/mwildt/progoter/request"
	"github.com/mwildt/progoter/response"
	"github.com/mwildt/progoter/tools"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ChatService struct {
	apiKey string
	client *http.Client
}

func NewChatService(apiKey string) *ChatService {
	return &ChatService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func readResponse(body io.Reader, messageChan chan *request.Message) (result *request.Message, err error) {
	reader := bufio.NewReader(body)
	result = &request.Message{}
	var builder strings.Builder
	var completition response.CompletionChunk

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

			// tool-calls
			if len(choice.Delta.ToolCalls) > 0 {
				result.ToolCalls = append(result.ToolCalls, choice.Delta.ToolCalls...)
			}
			if nil != messageChan {
				messageChan <- &request.Message{
					Role:      choice.Delta.Role,
					ToolCalls: choice.Delta.ToolCalls,
					Content:   contentPart,
				}
			}
		}
	}

	result.Content = builder.String()
	return result, nil
}

func (cs *ChatService) sendCompleteRequest(ctx context.Context, messages []*request.Message, messageChan chan *request.Message) (*request.Message, error) {
	jsonData, err := json.Marshal(&request.ChatCompletion{
		Model:    "devstral-medium-latest",
		Stream:   true,
		Messages: messages,
		Tools:    tools.GetTools(),
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

	slog.Default().Info("send completion request", "url", req.URL.String(), "jsonData", jsonData)

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

	return readResponse(resp.Body, messageChan)
}

// CompleteContext vervollständigt den ChatContext mit einer Antwort vom API.
func (cs *ChatService) CompleteContext(ctx context.Context, chatContext *ChatContext, messageChan chan *request.Message) (*ChatContext, error) {
	defer func() {
		if nil != messageChan {
			close(messageChan)
		}
	}()
	for {
		// wurde der context ggf in der zwischenzet abgebrochen?
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// // Prüfe, ob der Kontext zu mehr als 75% gefüllt ist
		/**
		if chatContext.TotalTokens > 0 && chatContext.TotalTokens > (262144*0.75) {
			slog.Default().Info("Kontext ist zu mehr als 75% gefüllt. Starte Compaction...")
			err := cs.compactContext(chatContext)
			if err != nil {
				slog.Error("Fehler bei der Compaction", "error", err)
				return nil, err
			}
		}
		*/

		responseMessage, err := cs.sendCompleteRequest(ctx, chatContext.GetMessages(), messageChan)
		if err != nil {
			return nil, err
		}
		chatContext.AddMessage(responseMessage)

		// Beende die Schleife, wenn keine Tool-Calls mehr anstehen
		if len(responseMessage.ToolCalls) == 0 {
			break
		} else {
			// tool calls ausführen und ggf weiter machen
			for _, toolCall := range responseMessage.ToolCalls {
				callContent, err := cs.callTool(toolCall)
				if err != nil {
					return nil, err
				}
				chatContext.AddMessage(&request.Message{
					Role:       "tool",
					ToolCallId: toolCall.Id,
					Content:    string(callContent),
				})
			}
		}
	}

	return chatContext, nil
}

// callTool ruft ein Tool auf und gibt das Ergebnis zurück.
func (cs *ChatService) callTool(call response.ToolCallChoice) ([]byte, error) {

	slog.Default().Info("Tool Call", "tool", call.Function.Name, "call_id", call.Id)

	if call.Function.Name == "read_file" {
		return tools.ReadFile(call.Function.Arguments)
	} else if call.Function.Name == "replace_file_content" {
		return tools.ReplaceFileContent(call.Function.Arguments)
	} else if call.Function.Name == "list_files" {
		return tools.ListFiles(call.Function.Arguments)
	} else if call.Function.Name == "write_file" {
		return tools.WriteFile(call.Function.Arguments)
	} else if call.Function.Name == "git_do" {
		return tools.GitDo(call.Function.Arguments)
	} else if call.Function.Name == "git_diff" {
		return tools.GitDiff(call.Function.Arguments)
	} else if call.Function.Name == "create_dir" {
		return tools.CreateDir(call.Function.Arguments)
	} else if call.Function.Name == "stop_process" {
		return tools.StopProcess(call.Function.Arguments)
	} else {
		return nil, errors.New("tool nicht gefunden")
	}
}
