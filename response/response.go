package response

type ChatCompletion struct {
	ID      string   `json:"id"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type Usage struct {
	PromptTokens        int           `json:"prompt_tokens"`
	TotalTokens         int           `json:"total_tokens"`
	CompletionTokens    int           `json:"completion_tokens"`
	PromptTokensDetails PromptDetails `json:"prompt_tokens_details"`
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
	Role         string           `json:"role"`
	Content      interface{}      `json:"content"`
	ToolCalls    []ToolCallChoice `json:"tool_calls"`
	FinishReason string           `json:"finish_reason"`
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

// :[{"index":0,"delta":{"tool_calls":[{"id":"IjcVstBlj","type":"function","function":{"name":"list_files","arguments":"{}"},"index":0}]},"finish_reason":"tool_calls"}]
