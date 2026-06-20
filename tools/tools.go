package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mwildt/progoter/request"
)

// GetTools liefert die Liste der verfügbaren Tools
func GetTools() []request.Tool {
	return []request.Tool{
		getWriteFileTool(),
		getListFilesTool(),
		getReadFileTool(),
		getReplaceFileContentTool(),
		getGitDoTool(),
		getGitDiffTool(),
		getCreateDirTool(),
		getStopProcessTool(),
	}
}

// WriteFileArgs enthält die Argumente für das write_file Tool
type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// getWriteFileTool definiert das write_file Tool
func getWriteFileTool() request.Tool {
	return request.Tool{
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
	}
}

// WriteFile implementiert das write_file Tool
func WriteFile(args string) ([]byte, error) {
	var writeFileArgs WriteFileArgs
	err := json.Unmarshal([]byte(args), &writeFileArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Schreibe den Inhalt in die Datei
	err = os.WriteFile(writeFileArgs.Path, []byte(writeFileArgs.Content), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Datei erfolgreich geschrieben oder erstellt"})
}

// getListFilesTool definiert das list_files Tool
func getListFilesTool() request.Tool {
	return request.Tool{
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
	}
}

// ListFiles implementiert das list_files Tool
func ListFiles(args string) ([]byte, error) {
	// Da keine Argumente erwartet werden, ignorieren wir den Eingabestring
	var files []string
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(files)
}

// ReadFileArgs enthält die Argumente für das read_file Tool
type ReadFileArgs struct {
	Path string `json:"path"`
}

// getReadFileTool definiert das read_file Tool
func getReadFileTool() request.Tool {
	return request.Tool{
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
	}
}

// ReadFile implementiert das read_file Tool
func ReadFile(args string) ([]byte, error) {
	var readFileArgs ReadFileArgs
	err := json.Unmarshal([]byte(args), &readFileArgs)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(readFileArgs.Path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// ReplaceFileContentArgs enthält die Argumente für das replace_file_content Tool
type ReplaceFileContentArgs struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// getReplaceFileContentTool definiert das replace_file_content Tool
func getReplaceFileContentTool() request.Tool {
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
	}
}

// ReplaceFileContent implementiert das replace_file_content Tool
func ReplaceFileContent(args string) ([]byte, error) {
	var replaceFileContentArgs ReplaceFileContentArgs
	err := json.Unmarshal([]byte(args), &replaceFileContentArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Überprüfe, ob die Datei existiert
	if _, err := os.Stat(replaceFileContentArgs.Path); os.IsNotExist(err) {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Lese den Inhalt der Datei
	content, err := ioutil.ReadFile(replaceFileContentArgs.Path)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Ersetze den Inhalt
	fileContent := string(content)
	var newFileContent string

	if replaceFileContentArgs.ReplaceAll {
		newFileContent = strings.ReplaceAll(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent)
	} else {
		newFileContent = strings.Replace(fileContent, replaceFileContentArgs.OldContent, replaceFileContentArgs.NewContent, 1)
	}

	// Schreibe den neuen Inhalt zurück in die Datei
	err = os.WriteFile(replaceFileContentArgs.Path, []byte(newFileContent), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Replacement erfolgreich"})
}

// GitDoArgs enthält die Argumente für das git_do Tool
type GitDoArgs struct {
	CommitMessage string `json:"commit_message"`
}

// getGitDoTool definiert das git_do Tool
func getGitDoTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "git_do",
			Description: "Fügt den aktuellen Stand zu Git hinzu, commited und pusht nach origin",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"commit_message": {
						Type:        "string",
						Name:        "commit_message",
						Description: "Die Commit-Nachricht für den Git-Commit.",
					},
				},
				Required: []string{
					"commit_message",
				},
			},
		},
	}
}

// GitDo implementiert das git_do Tool
func GitDo(args string) ([]byte, error) {
	var gitDoArgs GitDoArgs
	err := json.Unmarshal([]byte(args), &gitDoArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git add . aus
	if err := runCommand("git", "add", "."); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git commit aus
	if err := runCommand("git", "commit", "-m", gitDoArgs.CommitMessage); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git push aus
	if err := runCommand("git", "push", "origin"); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Git-Operationen erfolgreich ausgeführt"})
}

// getGitDiffTool definiert das git_diff Tool
func getGitDiffTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "git_diff",
			Description: "Zeigt die Änderungen seit dem letzten Commit an",
			Parameters: request.FunctionParams{
				Type:       "object",
				Properties: map[string]request.ArgumentProperty{},
				Required:   []string{},
			},
		},
	}
}

// GitDiff implementiert das git_diff Tool
func GitDiff(args string) ([]byte, error) {
	// Der Eingabestring wird ignoriert, da keine Argumente erwartet werden
	cmd := exec.Command("git", "diff", "HEAD")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Fehler beim Ausführen von git diff: %v (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// CreateDirArgs enthält die Argumente für das create_dir Tool
type CreateDirArgs struct {
	Path string `json:"path"`
}

// getCreateDirTool definiert das create_dir Tool
func getCreateDirTool() request.Tool {
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

// CreateDir implementiert das create_dir Tool
func CreateDir(args string) ([]byte, error) {
	var createDirArgs CreateDirArgs
	err := json.Unmarshal([]byte(args), &createDirArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	err = os.MkdirAll(createDirArgs.Path, 0755)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Verzeichnis erfolgreich erstellt"})
}

// StatusResponse enthält den Status und eine optionale Nachricht oder Fehler
type StatusResponse struct {
	Status   string `json:"status"`
	Messsage string `json:"messsage,omitempty"`
	Error    string `json:"error,omitempty"`
}

// StopProcessArgs enthält die Argumente für das stop_process Tool
type StopProcessArgs struct {
	PID int `json:"pid"`
}

// getStopProcessTool definiert das stop_process Tool
func getStopProcessTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "stop_process",
			Description: "Beendet einen laufenden Go-Prozess anhand seiner Prozess-ID",
			Parameters: request.FunctionParams{
				Type: "object",
				Properties: map[string]request.ArgumentProperty{
					"pid": {
						Type:        "integer",
						Name:        "pid",
						Description: "Die Prozess-ID des zu beendenden Prozesses.",
					},
				},
				Required: []string{
					"pid",
				},
			},
		},
	}
}

// StopProcess implementiert das stop_process Tool
func StopProcess(args string) ([]byte, error) {
	var stopProcessArgs StopProcessArgs
	err := json.Unmarshal([]byte(args), &stopProcessArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Beende den Prozess
	process, err := os.FindProcess(stopProcessArgs.PID)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	err = process.Kill()
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Prozess erfolgreich beendet"})
}

// runCommand führt einen Befehl aus
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Fehler beim Ausführen von %s: %v (stderr: %s)", name, err, stderr.String())
	}

	return nil
}
