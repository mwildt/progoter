package tools

import "github.com/mwildt/progoter/request"

type ToolDefinition request.Tool

// ToolHandler ist das Interface, das alle Tools implementieren müssen.
type ToolHandler interface {
	GetTool() ToolDefinition
	Execute(basePath string, args string) ([]byte, error)
}
