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
	"os"
	"strings"
)

type ChatService struct {
	apiKey      string
	client      *http.Client
	chatContext *ChatContext
	newMessages []*request.Message
}

func NewChatService(apiKey string) *ChatService {
	return &ChatService{
		apiKey:      apiKey,
		client:      &http.Client{},
		chatContext: NewChatContext(),
		newMessages: []*request.Message{},
	}
}

func (cs *ChatService) StartChat() {
	for {
		if len(cs.newMessages) == 0 {
			cs.chatContext.AddMessage(cs.getUserMessage("Was ist dein Begehr"))
		} else {
			cs.chatContext.AddMessages(cs.newMessages)
			cs.newMessages = []*request.Message{}
		}

		err := cs.processChatCompletion()
		if err != nil {
			slog.Error("Fehler beim Verarbeiten der Chat-Vervollständigung", "error", err)
		}
	}
}

func (cs *ChatService) processChatCompletion() error {
	jsonData, err := json.Marshal(&request.ChatCompletion{
		Model:    "devstral-medium-latest",
		Stream:   true,
		Messages: cs.chatContext.GetMessages(),
		Tools:    tools.GetTools(),
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	slog.Default().Info("send completion request", "url", req.URL.String(), "jsonData", jsonData)

	resp, err := cs.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Fehler beim Lesen des Response-Body", "error", err)
			resp.Body.Close()
			return err
		}

		resp.Body = io.NopCloser(bytes.NewBuffer(body))

		slog.Error("HTTP-Fehler", "status", resp.StatusCode, "body", string(body))
		resp.Body.Close()
		return errors.New("HTTP-Fehler")
	}

	reader := bufio.NewReader(resp.Body)

	var completition response.CompletionChunk

	responseMessage := &request.Message{
		Role: "assistant",
	}

	cs.chatContext.AddMessage(responseMessage)

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
				return err
			}

			first := completition.Choices[0]

			switch first.FinishReason {

			case "tool_calls":

				responseMessage.ToolCalls = append(responseMessage.ToolCalls, first.Delta.ToolCalls...)

				for _, toolCall := range first.Delta.ToolCalls {
					msg, err := cs.callTool(toolCall)
					if err != nil {
						cs.newMessages = append(cs.newMessages, &request.Message{
							Role:       "tool",
							ToolCallId: toolCall.Id,
							Content:    fmt.Sprintf("Beim Aufruf des Tools ist ein fehler aufgetreten. (error: %v)", err),
						})
					} else {
						cs.newMessages = append(cs.newMessages, &request.Message{
							Role:       "tool",
							ToolCallId: toolCall.Id,
							Content:    string(msg),
						})
					}
				}

			default:
				switch first.Delta.Content.(type) {
				case map[string]any:
					fmt.Printf("MAP %v\n", completition.Choices[0].Delta.Content)
				case []any:
					fmt.Printf("LIST %v\n", completition.Choices[0].Delta.Content)
				case string:
					fmt.Print(completition.Choices[0].Delta.Content.(string))
					builder.WriteString(completition.Choices[0].Delta.Content.(string))

				default:
					fmt.Printf("Unknown type %v\n", completition.Choices[0].Delta.Content)
				}

			}

		}
	}

	if builder.Len() > 0 {
		println()
		responseMessage.Content = builder.String()
	}

	resp.Body.Close()

	return nil
}

func (cs *ChatService) getUserMessage(s string) *request.Message {
	reader := bufio.NewReader(os.Stdin)
	println(s)
	input, _ := reader.ReadString('\n')
	return &request.Message{
		Role:    "user",
		Content: input,
	}
}

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
