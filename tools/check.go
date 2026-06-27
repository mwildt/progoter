package tools

import (
	"bytes"
	"fmt"
	"github.com/mwildt/progoter/request"
	"io"
	"os"
	"os/exec"
)

// CheckTool implementiert das ToolHandler-Interface für check
type CheckTool struct{}

func (t CheckTool) GetTool() request.Tool {
	return request.Tool{
		Type: "function",
		Function: request.ToolFunction{
			Name:        "check",
			Description: "Startet einen Podman-Container mit dem letzten Go-Image und versucht, alle main.go-Dateien zu bauen.",
			Parameters: request.FunctionParams{
				Type:       "object",
				Properties: map[string]request.ArgumentProperty{},
				Required:   []string{},
			},
		},
	}
}

const src = `
#!/bin/bash

find . -name "main.go" -exec sh -c '
    dir=$(dirname "$1")
    echo "Building in $dir..."

    if (cd "$dir" && go build -o /dev/null .); then
        echo "Build successful in $dir"
    else
        echo "Build failed in $dir"
    fi

    echo "----------------------------------------"
' sh {} \;
`

func (t CheckTool) Execute(basePath string, args string) ([]byte, error) {
	// Führe podman run aus, um einen Container mit dem Go-Image zu starten
	// und alle main.go-Dateien zu bauen
	cmd := exec.Command("podman", "run",
		"--rm",
		"-w", "/workspace",
		"-v", fmt.Sprintf("%s:/workspace:ro", basePath),
		"docker.io/library/golang:latest",
		"bash", "-c", src)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Fehler beim Ausführen des Podman-Containers: %v (stderr: %s)", err, stderrBuf.String())
	}

	return stdoutBuf.Bytes(), nil
}
