package response

import "github.com/mwildt/progoter/request"

type ToolCallChoice request.ToolCallChoice
type Usage request.Usage

type ChatCompletion struct {
	ID      string   `json:"id"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type PromptDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type Choice struct {
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason"`
	Message      ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Role      string `json:"role"`
	ToolCalls any    `json:"tool_calls"`
	Content   string `json:"content"`
}

type CompletionChunk struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []CompletionChunkChoice `json:"choices"`
	Usage   Usage                   `json:"usage"`
}

type CompletionChunkChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type Delta struct {
	Role         string                   `json:"role"`
	Content      interface{}              `json:"content"`
	ToolCalls    []request.ToolCallChoice `json:"tool_calls"`
	FinishReason string                   `json:"finish_reason"`
}
