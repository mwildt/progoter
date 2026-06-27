package tools

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/mwildt/progoter/request"
)

// ReadFileArgs enthält die Argumente für das read_file Tool
type ReadFileArgs struct {
	Path string `json:"path"`
}

// ReadFileTool implementiert das ToolHandler-Interface für read_file
type ReadFileTool struct{}

func (t ReadFileTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "read_file",
			Description: "Liest den Inhalt einer Datei",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"path": {
						Type:        "string",
						Name:        "path",
						Description: "Path",
					},
				},
				Required: []string{
					"path",
				},
			},
		},
	}
}

func (t ReadFileTool) Execute(basePath string, args string) ([]byte, error) {
	var readFileArgs ReadFileArgs
	err := json.Unmarshal([]byte(args), &readFileArgs)
	if err != nil {
		return nil, err
	}

	finalPath, err := filepath.Abs(path.Join(basePath, readFileArgs.Path))
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(finalPath)
	if err != nil {
		return nil, err
	}
	return content, nil
}
