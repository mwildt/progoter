package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mwildt/progoter/request"
)

// ReplaceFileContentArgs enthält die Argumente für das replace_file_content Tool
type ReplaceFileContentArgs struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// ReplaceFileContentTool implementiert das ToolHandler-Interface für replace_file_content
type ReplaceFileContentTool struct{}

func (t ReplaceFileContentTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "replace_file_content",
			Description: "Ersetzt einen Teil des Inhalts einer Datei",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"path": {
						Type:        "string",
						Name:        "path",
						Description: "Pfad zur Datei, die bearbeitet werden soll.",
					},
					"old_content": {
						Type:        "string",
						Name:        "old_content",
						Description: "Der Text, der ersetzt werden soll.",
					},
					"new_content": {
						Type:        "string",
						Name:        "new_content",
						Description: "Der neue Text, der den alten ersetzen soll.",
					},
					"replace_all": {
						Type:        "boolean",
						Name:        "replace_all",
						Description: "Wenn true, werden alle Vorkommen von old_content ersetzt. Andernfalls nur das erste.",
					},
				},
				Required: []string{
					"path", "new_content", "old_content",
				},
			},
		},
	}
}

func (t ReplaceFileContentTool) Execute(basePath string, args string) ([]byte, error) {
	var replaceFileContentArgs ReplaceFileContentArgs
	err := json.Unmarshal([]byte(args), &replaceFileContentArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	finalPath, err := filepath.Abs(path.Join(basePath, replaceFileContentArgs.Path))
	if err != nil {
		return nil, err
	}

	// Überprüfe, ob die Datei existiert
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Lese den Inhalt der Datei
	content, err := os.ReadFile(finalPath)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	fileContent := string(content)

	// Prüfe, ob der zu ersetzende Inhalt vorhanden ist
	if !strings.Contains(fileContent, replaceFileContentArgs.OldContent) {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: "der zu ersetzende Inhalt wurde in der Datei nicht gefunden"})
		return status, fmt.Errorf("der zu ersetzende Inhalt wurde in der Datei nicht gefunden")
	}

	var newFileContent string
	if replaceFileContentArgs.ReplaceAll {
		newFileContent = strings.ReplaceAll(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent)
	} else {
		newFileContent = strings.Replace(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent, 1)
	}

	// Schreibe den neuen Inhalt zurück in die Datei
	err = os.WriteFile(finalPath, []byte(newFileContent), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Replacement erfolgreich"})
}
