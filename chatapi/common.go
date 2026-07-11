package chatapi

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
