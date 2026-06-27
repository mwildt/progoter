package tools

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/mwildt/progoter/request"
)

// ListFilesArgs enthält die Argumente für das list_files Tool
type ListFilesArgs struct {
	Pattern string `json:"pattern,omitempty"`
}

// ListFilesTool implementiert das ToolHandler-Interface für list_files
type ListFilesTool struct{}

func (t ListFilesTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "list_files",
			Description: "Gibt eine Liste mit allen Dateien im Projekt zurück, die einem Glob-Ausdruck entsprechen",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"pattern": {
						Type:        "string",
						Name:        "pattern",
						Description: "Ein optionaler Glob-Ausdruck, um Dateien zu filtern. Wenn nicht angegeben, werden alle Dateien zurückgegeben.",
					},
				},
				Required: []string{},
			},
		},
	}
}

func (t ListFilesTool) Execute(basePath string, args string) ([]byte, error) {
	var listFilesArgs ListFilesArgs
	err := json.Unmarshal([]byte(args), &listFilesArgs)
	if err != nil {
		return nil, err
	}

	finalPath, err := filepath.Abs(path.Join(basePath, "./"))
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(finalPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			relPath, err := filepath.Rel(finalPath, path)
			if err != nil {
				return err
			}

			// Wenn ein Glob-Muster angegeben wurde, prüfe, ob die Datei dazu passt
			if listFilesArgs.Pattern != "" {
				matched, err := filepath.Match(listFilesArgs.Pattern, relPath)
				if err != nil {
					return err
				}
				if matched {
					files = append(files, relPath)
				}
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(files)
}
