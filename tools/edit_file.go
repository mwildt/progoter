package tools

import (
	"encoding/json"
	"github.com/mwildt/progoter/request"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReplaceFileContentArgs enthält die Argumente für das replace_file_content Tool
type ReplaceFileContentArgs struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// StatusResponse enthält den Status und eine Nachricht oder Fehler

// EditFileTool implementiert das ToolHandler-Interface für replace_file_content
type EditFileTool struct {
}

func (t EditFileTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "edit_file",
			Description: "Ersetzt einen Teil des Inhalts einer Datei. Der zu ersetzende Teil muss byte-genau angegeben werden.",
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

func (t EditFileTool) Execute(basePath string, args string) ([]byte, error) {
	var replaceFileContentArgs ReplaceFileContentArgs
	err := json.Unmarshal([]byte(args), &replaceFileContentArgs)
	if err != nil {
		return t.errorResponse("Fehler beim Parsen der Argumente: " + err.Error())
	}

	// Validierung der Argumente
	if replaceFileContentArgs.OldContent == replaceFileContentArgs.NewContent {
		return t.errorResponse("old_content und new_content sind identisch. Keine Änderungen erforderlich.")
	}

	if replaceFileContentArgs.OldContent == "" {
		return t.errorResponse("old_content darf nicht leer sein.")
	}

	finalPath, err := filepath.Abs(path.Join(basePath, replaceFileContentArgs.Path))
	if err != nil {
		return t.errorResponse("Fehler beim Erstellen des absoluten Pfads: " + err.Error())
	}

	// Überprüfe, ob die Datei existiert
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return t.errorResponse("Die Datei existiert nicht: " + finalPath)
	}

	// Lese den Inhalt der Datei
	content, err := os.ReadFile(finalPath)
	if err != nil {
		return t.errorResponse("Fehler beim Lesen der Datei: " + err.Error())
	}

	fileContent := string(content)

	// Prüfe, ob der zu ersetzende Inhalt vorhanden ist
	if !strings.Contains(fileContent, replaceFileContentArgs.OldContent) {
		return t.errorResponse("Der zu ersetzende Inhalt wurde in der Datei nicht gefunden.")
	}

	// Ersetze den Inhalt
	var newFileContent string
	if replaceFileContentArgs.ReplaceAll {
		newFileContent = strings.ReplaceAll(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent)
	} else {
		newFileContent = strings.Replace(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent, 1)
	}

	err = os.WriteFile(finalPath, []byte(newFileContent), 0644)
	if err != nil {
		return t.errorResponse("Fehler beim Schreiben der Datei: " + err.Error())
	}

	return t.successResponse("Replacement erfolgreich")
}

func (t *EditFileTool) errorResponse(message string) ([]byte, error) {
	status := StatusResponse{Status: "ERROR", Error: message}
	return json.Marshal(status)
}

func (t *EditFileTool) successResponse(message string) ([]byte, error) {
	status := StatusResponse{Status: "OK", Message: message}
	return json.Marshal(status)
}
