package tools

import (
	"github.com/mwildt/progoter/request"
)

// GetTools liefert die Liste der verfügbaren Tools
func GetTools() []request.Tool {
	return []request.Tool{
		{
			Type: "function",
			Function: request.ToolFunction{
				Name:        "write_file",
				Description: "Schreibt den Inhalt in eine Datei oder erstellt eine neue Datei",
				Parameters: request.FunctionParams{
					Type: "object",
					Properties: map[string]request.ArgumentProperty{
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
		},
		{
			Type: "function",
			Function: request.ToolFunction{
				Name:        "list_files",
				Description: "Gibt eine Liste mit allen Dateien im Projekt zurück",
				Parameters: request.FunctionParams{
					Type:       "object",
					Properties: map[string]request.ArgumentProperty{},
					Required:   []string{},
				},
			},
		},
		{
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
		},
		{
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
							Name:        "Pfad zur Datei, die bearbeitet werden soll.",
							Description: "Path",
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
		},
	}
}
