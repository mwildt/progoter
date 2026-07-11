package tools

import (
	"github.com/mwildt/progoter/chatapi"
)

type ToolDefinition chatapi.Tool

// ToolHandler ist das Interface, das alle Tools implementieren müssen.
type ToolHandler interface {
	GetTool() ToolDefinition
	Execute(basePath string, args string) ([]byte, error)
}
