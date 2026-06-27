package tools

import (
	"os"

	"github.com/mwildt/progoter/request"
)

// StopProcessTool implementiert das ToolHandler-Interface für stop_process
type StopProcessTool struct{}

func (t StopProcessTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "stop_process",
			Description: "Beendet den laufenden Prozess.",
			Parameters: request.FunctionParams{
				Type:       "object",
				Properties: map[string]request.ArgumentProperty{},
				Required:   []string{},
			},
		},
	}
}

func (t StopProcessTool) Execute(basePath string, args string) ([]byte, error) {
	os.Exit(1)
	return nil, nil
}
