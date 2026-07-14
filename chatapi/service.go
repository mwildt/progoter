package chatapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type MessageHandler interface {
	Join(*Message)
}

type MessageHandlerFunc func(*Message)

func (fn MessageHandlerFunc) Join(message *Message) {
	fn(message)
}

type Service struct {
	baseUrl string // https://api.mistral.ai/v1/
	apiKey  string
	client  *http.Client
}

func NewService(apiKey string) *Service {
	return &Service{
		baseUrl: "https://api.mistral.ai/v1/",
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

func (service *Service) Complete(ctx context.Context, request ChatCompletionRequest, handler MessageHandler) (*Message, error) {

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", service.baseUrl+"chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	slog.Default().With("logger", "ChatService").
		Info("send completion request", "url", req.URL.String())

	resp, err := service.client.Do(req)
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

	return service.readResponse(resp.Body, handler)
}

func (service *Service) readResponse(body io.Reader, handler MessageHandler) (result *Message, err error) {
	reader := bufio.NewReader(body)
	result = &Message{}
	var builder strings.Builder
	var completition CompletionChunk

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

			result.Usage = Usage(completition.Usage)

			// tool-calls
			if len(choice.Delta.ToolCalls) > 0 {
				result.ToolCalls = append(result.ToolCalls, choice.Delta.ToolCalls...)
			}
			if nil != handler {
				handler.Join(&Message{
					Role:      choice.Delta.Role,
					ToolCalls: choice.Delta.ToolCalls,
					Content:   contentPart,
					Usage:     Usage(completition.Usage),
				})
			}
		}
	}

	slog.Default().Info("EOD")

	result.Content = builder.String()
	return result, nil
}
