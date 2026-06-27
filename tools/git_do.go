package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/mwildt/progoter/request"
)

// GitDoArgs enthält die Argumente für das git_do Tool
type GitDoArgs struct {
	CommitMessage string `json:"commit_message"`
}

// GitDoTool implementiert das ToolHandler-Interface für git_do
type GitDoTool struct{}

func (t GitDoTool) GetTool() request.Tool {
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

func (t GitDoTool) Execute(basePath string, args string) ([]byte, error) {
	var gitDoArgs GitDoArgs
	err := json.Unmarshal([]byte(args), &gitDoArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git add . aus
	if err := runCommand("git", "-C", basePath, "add", "."); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git commit aus
	if err := runCommand("git", "-C", basePath, "commit", "-m", gitDoArgs.CommitMessage); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git push aus
	if err := runCommand("git", "-C", basePath, "push", "origin"); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Git-Operationen erfolgreich ausgeführt"})
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
