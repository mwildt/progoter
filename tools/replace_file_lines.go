package tools

import (
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReplaceFileLinesArgs enthält die Argumente für das replace_file_lines Tool
type ReplaceFileLinesArgs struct {
	Path       string `json:"path"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	NewContent string `json:"new_content"`
}

// ReplaceFileLinesTool implementiert das ToolHandler-Interface für replace_file_lines
type ReplaceFileLinesTool struct{}

func (t ReplaceFileLinesTool) GetTool() ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "replace_file_lines",
			Description: "Ersetzt einen Bereich von Zeilen in einer Datei. ",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
					"path": {
						Type:        "string",
						Name:        "path",
						Description: "Pfad zur Datei, die bearbeitet werden soll.",
					},
					"start_line": {
						Type:        "integer",
						Name:        "start_line",
						Description: "Die Startzeile, ab der ersetzt werden soll (inklusiv).",
					},
					"end_line": {
						Type:        "integer",
						Name:        "end_line",
						Description: "Die Endzeile, bis zu der ersetzt werden soll (inklusiv).",
					},
					"new_content": {
						Type:        "string",
						Name:        "new_content",
						Description: "Der neue Inhalt, der die Zeilen ersetzen soll.",
					},
				},
				Required: []string{
					"path", "start_line", "end_line", "new_content",
				},
			},
		},
	}
}

func (t ReplaceFileLinesTool) Execute(basePath string, args string) ([]byte, error) {
	var replaceFileLinesArgs ReplaceFileLinesArgs
	err := json.Unmarshal([]byte(args), &replaceFileLinesArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	finalPath, err := filepath.Abs(path.Join(basePath, replaceFileLinesArgs.Path))
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
	lines := strings.Split(fileContent, "\n")

	// Überprüfe, ob die Zeilenangaben gültig sind
	if replaceFileLinesArgs.StartLine < 1 || replaceFileLinesArgs.EndLine < 1 ||
		replaceFileLinesArgs.StartLine > len(lines) || replaceFileLinesArgs.EndLine > len(lines) {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: "ungültige Zeilenangaben"})
		return status, fmt.Errorf("ungültige Zeilenangaben")
	}

	// Ersetze die Zeilen
	newLines := append(lines[:replaceFileLinesArgs.StartLine-1], strings.Split(replaceFileLinesArgs.NewContent, "\n")...)
	newLines = append(newLines, lines[replaceFileLinesArgs.EndLine:]...)

	// Schreibe den neuen Inhalt zurück in die Datei
	err = os.WriteFile(finalPath, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Message: "Zeilenersetzung erfolgreich"})
}
