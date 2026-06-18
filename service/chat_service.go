package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

// CompleteContext vervollständigt den ChatContext mit einer Antwort vom API.
func (cs *ChatService) CompleteContext(chatContext *ChatContext, messageChan chan *request.Message) (*ChatContext, error) {
	defer func() {
		if nil != messageChan {
			close(messageChan)
		}
	}()
	for {
		jsonData, err := json.Marshal(&request.ChatCompletion{
			Model:    "devstral-medium-latest",
			Stream:   true,
			Messages: chatContext.GetMessages(),
			Tools:    tools.GetTools(),
		})

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
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

		reader := bufio.NewReader(resp.Body)

		var completition response.CompletionChunk

		responseMessage := &request.Message{
			Role: "assistant",
		}

		chatContext.AddMessage(responseMessage)

		var builder strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			line = strings.TrimSpace(line)

			// SSE Format: "data: ..."
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)

				if data == "[DONE]" {
					break
				}

				err := json.Unmarshal([]byte(data), &completition)
				if err != nil {
					return nil, err
				}

				first := completition.Choices[0]

				switch first.FinishReason {

				case "tool_calls":

					responseMessage.ToolCalls = append(responseMessage.ToolCalls, first.Delta.ToolCalls...)

					for _, toolCall := range first.Delta.ToolCalls {
						// Sende den Tool-Call an den messageChan, falls dieser nicht nil ist
						if messageChan != nil {
							messageChan <- &request.Message{
								Role:    "tool-request",
								Content: fmt.Sprintf("Tool-Call: %s mit ID %s", toolCall.Function.Name, toolCall.Id),
							}
						}

						msg, err := cs.callTool(toolCall)
						if err != nil {
							chatContext.AddMessage(&request.Message{
								Role:       "tool",
								ToolCallId: toolCall.Id,
								Content:    fmt.Sprintf("Beim Aufruf des Tools ist ein fehler aufgetreten. (error: %v)", err),
							})
							// Sende die Fehlermeldung an den messageChan, falls dieser nicht nil ist
							if messageChan != nil {
								messageChan <- &request.Message{
									Role:    "tool",
									Content: fmt.Sprintf("Fehler beim Aufruf des Tools: %v", err),
								}
							}
						} else {
							chatContext.AddMessage(&request.Message{
								Role:       "tool",
								ToolCallId: toolCall.Id,
								Content:    string(msg),
							})
							// Sende die Tool-Antwort an den messageChan, falls dieser nicht nil ist
							if messageChan != nil {
								messageChan <- &request.Message{
									Role:    "tool",
									Content: string(msg),
								}
							}
						}
					}

				default:
					switch content := first.Delta.Content.(type) {
					case map[string]any:
						fmt.Printf("MAP %v\n", content)
					case []any:
						fmt.Printf("LIST %v\n", content)
					case string:
						if messageChan != nil {
							messageChan <- &request.Message{
								Role:    "assistant",
								Content: content,
							}
						}
						builder.WriteString(content)

					default:
						fmt.Printf("Unknown type %v\n", content)
					}
				}
				chatContext.TotalTokens = completition.Usage.TotalTokens
			}
		}

		if builder.Len() > 0 {
			println()
			responseMessage.Content = builder.String()
		}

		resp.Body.Close()

		// Beende die Schleife, wenn keine Tool-Calls mehr anstehen
		if len(responseMessage.ToolCalls) == 0 {
			break
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
	} else {
		return nil, errors.New("tool nicht gefunden")
	}
}
