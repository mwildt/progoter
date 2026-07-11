package chatapi

type ChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Stream   bool                     `json:"stream"`
	Messages []*ChatCompletionMessage `json:"messages"`
	Tools    []Tool                   `json:"tools,omitempty"`
}

type ChatCompletionMessage struct {
	Role       string           `json:"role"`
	ToolCallId string           `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallChoice `json:"tool_calls,omitempty"`
	Content    string           `json:"content,omitempty"`
}

type Tool struct {
	Type     string       `json:"type"` // z.B. "function"
	Function ToolFunction `json:"function"`
}

type ArgumentProperty struct {
	Type        string            `json:"type"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Items       *ArgumentProperty `json:"items,omitempty"`
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
