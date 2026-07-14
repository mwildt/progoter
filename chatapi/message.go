package chatapi

type Message struct {
	Role       string           `json:"role"`
	ToolCallId string           `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallChoice `json:"tool_calls,omitzero"`
	Content    string           `json:"content,omitempty"`
	Usage      Usage            `json:"usage,omitzero"`
}

func FromMessage(m *Message) *Message {
	return &Message{
		Role:       m.Role,
		ToolCallId: m.ToolCallId,
		ToolCalls:  m.ToolCalls,
		Content:    m.Content,
		Usage:      m.Usage,
	}
}

func (m *Message) HasRole(role string) bool {
	return m.Role == role
}

func (m *Message) Join(message *Message) {
	m.Content += message.Content
	m.ToolCalls = append(m.ToolCalls, message.ToolCalls...)
	m.Usage = message.Usage
}
