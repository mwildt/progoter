package tools

// TESTKOMMENTAR
import (
	"bufio"
	"encoding/json"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
)

type SearchInFilesTool struct {
	Exclusions FileExclusions
}

type SearchInFilesArgs struct {
	Pattern string `json:"pattern,omitempty"`
}

func (t SearchInFilesTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "search_in_files",
			Description: "Durchsucht alle Dateien im angegebenen Pfad nach einem Wort oder Regex-Muster.",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"pattern": {
						Type:        "string",
						Name:        "pattern",
						Description: "Das Wort oder Muster, nach dem gesucht werden soll. Hier ist Regex Möglich!",
					},
				},
				Required: []string{"pattern"},
			},
		},
	}
}

// SearchInFiles durchsucht alle Dateien im angegebenen Pfad nach einem Muster.
type SearchResult struct {
	FilePath string      `json:"path"`
	Matches  []FileMatch `json:"matches"`
}

type FileMatch struct {
	Line  int    `json:"line"`
	Match string `json:"match"`
}

func (t SearchInFilesTool) handleFile(path string, matchLine func(line string) (bool, string)) (match bool, matches []FileMatch, err error) {
	file, err := os.Open(path)
	if err != nil {
		return match, matches, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if lineMatch, value := matchLine(scanner.Text()); lineMatch {
			match = true
			matches = append(matches, FileMatch{
				Line:  lineNumber,
				Match: value,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Default().Error("Error beim scaning", "file", path, "error", err)
		return match, matches, err
	}
	return match, matches, nil
}

func (t SearchInFilesTool) Execute(basePath string, args string) ([]byte, error) {
	var searchArgs SearchInFilesArgs
	err := json.Unmarshal([]byte(args), &searchArgs)
	if err != nil {
		return errorResponse("Fehler beim Parsen der Argumente", err)
	}

	finalPath, err := filepath.Abs(basePath)
	if err != nil {
		return errorResponse("Fehler beim Erstellen des Pfads", err)
	}

	regex, err := regexp.Compile(searchArgs.Pattern)
	if err != nil {
		return errorResponse("Fehler beim Kompilieren des Regex-Musters", err)
	}

	lineMatcher := func(line string) (bool, string) {
		matches := regex.FindStringSubmatch(line)
		if matches != nil {
			return true, matches[0]
		} else {
			return false, ""
		}
	}

	var results []SearchResult

	err = filepath.Walk(finalPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(finalPath, filePath)
		if t.Exclusions.Match(relPath) {
			return filepath.SkipDir
		}
		if match, matches, err := t.handleFile(filePath, lineMatcher); err != nil {
			return err
		} else if match {
			result := SearchResult{
				FilePath: relPath,
				Matches:  matches,
			}
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return errorResponse("Fehler beim Durchlaufen des Pfads", err)
	}

	slog.Default().Info("Suche abgeschlossen", "results", results)
	return json.Marshal(results)
}
