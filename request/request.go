package request

import "github.com/mwildt/progoter/response"

type Message struct {
	Role       string                    `json:"role"`
	ToolCallId string                    `json:"tool_call_id,omitempty"`
	ToolCalls  []response.ToolCallChoice `json:"tool_calls,omitempty"`
	Content    string                    `json:"content,omitempty"`
}

type Tool struct {
	Type     string       `json:"type"` // z.B. "function"
	Function ToolFunction `json:"function"`
}

type ArgumentProperty struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FunctionParams struct {
	Type       string                      `json:"type"` // z.B. "object"
	Properties map[string]ArgumentProperty `json:"properties"`
	Required   []string                    `json:"required,omitempty"`
}

type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  FunctionParams `json:"parameters"`
}

type ChatCompletion struct {
	Model    string     `json:"model"`
	Stream   bool       `json:"stream"`
	Messages []*Message `json:"messages"`
	Tools    []Tool     `json:"tools,omitempty"`
}
