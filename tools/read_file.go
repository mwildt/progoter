package tools

import (
	"encoding/json"
	"github.com/mwildt/progoter/chatapi"
	"os"
	"path"
	"path/filepath"
)

// ReadFileArgs enthält die Argumente für das read_file Tool
type ReadFileArgs struct {
	Path string `json:"path"`
}

// ReadFileTool implementiert das ToolHandler-Interface für read_file
type ReadFileTool struct{}

func (t ReadFileTool) GetTool() ToolDefinition {
	return ToolDefinition(chatapi.Tool{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "read_file",
			Description: "Liest den Inhalt einer Datei",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
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
	})
}

func (t ReadFileTool) Execute(basePath string, args string) ([]byte, error) {
	var readFileArgs ReadFileArgs
	err := json.Unmarshal([]byte(args), &readFileArgs)
	if err != nil {
		return errorResponse("Args konnten nicht geparst werden", err)
	}

	finalPath, err := filepath.Abs(path.Join(basePath, readFileArgs.Path))
	if err != nil {
		return errorResponse("Probleme mit dem FilePath", err)
	}

	content, err := os.ReadFile(finalPath)
	if err != nil {
		return errorResponse("Datei konnte nicht gelesen werden", err)
	}
	return content, nil
}
