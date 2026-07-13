package tools

import (
	"bytes"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"os/exec"
)

// GitDiffTool implementiert das ToolHandler-Interface für git_diff
type GitDiffTool struct{}

func (t GitDiffTool) GetTool() ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "git_diff",
			Description: "Zeigt die Änderungen seit dem letzten Commit an",
			Parameters: chatapi.FunctionParams{
				Type:       "object",
				Properties: map[string]chatapi.ArgumentProperty{},
				Required:   []string{},
			},
		},
	}
}

func (t GitDiffTool) Execute(basePath string, args string) ([]byte, error) {
	// Der Eingabestring wird ignoriert, da keine Argumente erwartet werden
	cmd := exec.Command("git", "-C", basePath, "diff", "HEAD")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Fehler beim Ausführen von git diff: %v (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
