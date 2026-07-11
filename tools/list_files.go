package tools

import (
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"github.com/mwildt/progoter/utils/glob"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ListFilesTool struct {
	excludeDirs []string
}

type ListFilesArgs struct {
	Pattern string `json:"pattern,omitempty"`
}

func (t ListFilesTool) GetTool() chatapi.Tool {
	return chatapi.Tool{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "list_files",
			Description: "Gibt eine Liste mit allen Dateien im Projekt zurück, die einem Glob-Ausdruck entsprechen. Denke an WildCards (*) beim aufruf. ",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
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

// ListFiles lists all files matching the given glob pattern, excluding specified directories.
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

	var matches []string

	glob := glob.NewGlob(listFilesArgs.Pattern)

	// Walk through the directory tree
	err = filepath.Walk(finalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(finalPath, path)

		if strings.HasPrefix(relPath, ".git/") {
			return filepath.SkipDir
		}

		if strings.HasPrefix(relPath, ".idea/") {
			return filepath.SkipDir
		}

		// Skip directories in excludeDirs
		for _, dir := range t.excludeDirs {
			if strings.HasPrefix(relPath, dir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if listFilesArgs.Pattern != "" && !glob.Match(relPath) {
			return nil
		}

		if info.IsDir() {
			matches = append(matches, relPath+"/")
		} else {
			matches = append(matches, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path: %v", err)
	}
	slog.Default().Info("list file respons", "matches", matches)
	return json.Marshal(matches)
}
