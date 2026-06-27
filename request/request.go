package request

type Message struct {
	Role       string           `json:"role"`
	ToolCallId string           `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallChoice `json:"tool_calls,omitzero"`
	Content    string           `json:"content,omitempty"`
	Usage      Usage            `json:"usage,omitzero"`
}

type ToolCallChoice struct {
	Id       string       `json:"id"`
	Index    int          `json:"index"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
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
