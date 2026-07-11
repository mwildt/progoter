package tools

import (
	"github.com/mwildt/progoter/chatapi"
	"os"
)

// StopProcessTool implementiert das ToolHandler-Interface für stop_process
type StopProcessTool struct{}

func (t StopProcessTool) GetTool() chatapi.Tool {
	return chatapi.Tool{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "stop_process",
			Description: "Beendet den laufenden Prozess.",
			Parameters: chatapi.FunctionParams{
				Type:       "object",
				Properties: map[string]chatapi.ArgumentProperty{},
				Required:   []string{},
			},
		},
	}
}

func (t StopProcessTool) Execute(basePath string, args string) ([]byte, error) {
	os.Exit(1)
	return nil, nil
}
