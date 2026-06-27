package tools

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/mwildt/progoter/request"
)

// CreateDirArgs enthält die Argumente für das create_dir Tool
type CreateDirArgs struct {
	Path string `json:"path"`
}

// CreateDirTool implementiert das ToolHandler-Interface für create_dir
type CreateDirTool struct{}

func (t CreateDirTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "create_dir",
			Description: "Erstellt ein neues Verzeichnis",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"path": {
						Type:        "string",
						Name:        "path",
						Description: "Pfad zum Verzeichnis, das erstellt werden soll.",
					},
				},
				Required: []string{
					"path",
				},
			},
		},
	}
}

func (t CreateDirTool) Execute(basePath string, args string) ([]byte, error) {
	var createDirArgs CreateDirArgs
	err := json.Unmarshal([]byte(args), &createDirArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	finalPath, err := filepath.Abs(path.Join(basePath, createDirArgs.Path))
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(finalPath, 0755)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Verzeichnis erfolgreich erstellt"})
}
