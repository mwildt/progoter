package tools

import (
	"encoding/json"
	"github.com/mwildt/progoter/chatapi"
	"os"
	"path"
	"path/filepath"
)

// WriteFileArgs enthält die Argumente für das write_file Tool
type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileTool implementiert das ToolHandler-Interface für write_file
type WriteFileTool struct{}

func (t WriteFileTool) GetTool() ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "write_file",
			Description: "Schreibt den Inhalt in eine Datei oder erstellt eine neue Datei",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
					"path": {
						Type:        "string",
						Name:        "path",
						Description: "Pfad zur Datei, die geschrieben oder erstellt werden soll.",
					},
					"content": {
						Type:        "string",
						Name:        "content",
						Description: "Der Inhalt, der in die Datei geschrieben werden soll.",
					},
				},
				Required: []string{
					"path", "content",
				},
			},
		},
	}
}

func (t WriteFileTool) Execute(basePath string, args string) ([]byte, error) {
	var writeFileArgs WriteFileArgs
	err := json.Unmarshal([]byte(args), &writeFileArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	finalPath, err := filepath.Abs(path.Join(basePath, writeFileArgs.Path))
	if err != nil {
		return nil, err
	}

	// Schreibe den Inhalt in die Datei
	err = os.WriteFile(finalPath, []byte(writeFileArgs.Content), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	return json.Marshal(StatusResponse{Status: "OK", Message: "Datei erfolgreich geschrieben oder erstellt"})
}
