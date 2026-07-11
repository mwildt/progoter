package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/chatapi"
	"io"
	"os"
	"os/exec"
)

// GolangTool implementiert das ToolHandler-Interface für golang
type GolangTool struct{}

func (t GolangTool) GetTool() chatapi.Tool {
	return chatapi.Tool{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "golang",
			Description: "Führt einen beliebigen Go-Befehl in einem Podman-Container aus.",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
					"command": {
						Type:        "array",
						Description: "Die Argumente für den Go-Befehl, z. B. ['test', './...'] oder ['run', 'main.go'].",
						Items: &chatapi.ArgumentProperty{
							Type: "string",
						},
					},
				},
				Required: []string{"command"},
			},
		},
	}
}

type GolangToolArgs struct {
	Command []string `json:"command"`
}

func (t GolangTool) Execute(basePath string, args string) ([]byte, error) {
	var toolArgs GolangToolArgs
	err := json.Unmarshal([]byte(args), &toolArgs)
	if err != nil {
		return []byte("Fehler"), err
	}

	cmdArgs := []string{
		"run",
		"--rm",
		"-w", "/workspace",
		"-v", fmt.Sprintf("%s:/workspace", basePath),
		"docker.io/library/golang:latest",
		"go",
	}

	cmdArgs = append(cmdArgs, toolArgs.Command...)
	cmd := exec.Command("podman", cmdArgs...)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Run(); err != nil {
		return errorResponse(stderrBuf.String(), err)
	}

	return successResponse(stdoutBuf.String())
}
