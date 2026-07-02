package tools

import (
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/request"
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

func (t ListFilesTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "list_files",
			Description: "Gibt eine Liste mit allen Dateien im Projekt zurück, die einem Glob-Ausdruck entsprechen. Denke an WildCards (*) beim aufruf. ",
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

	pattern := listFilesArgs.Pattern
	// If no pattern is provided, use "*" to match all files
	if pattern == "" {
		pattern = "*"
	}

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

		slog.Default().Info("check file info", "path", path, "relPath", relPath, "pattern", pattern)

		// Skip directories in excludeDirs
		for _, dir := range t.excludeDirs {
			if strings.HasPrefix(relPath, dir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if the path matches the pattern
		matched, err := filepath.Match(pattern, relPath)
		if err != nil {
			return err
		}

		if matched {
			if info.IsDir() {
				matches = append(matches, relPath+"/")
			} else {
				matches = append(matches, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path: %v", err)
	}

	return json.Marshal(matches)
}
