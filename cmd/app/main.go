package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mwildt/progoter/request"
	"github.com/mwildt/progoter/response"
	"github.com/mwildt/progoter/tools"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	// Lade die .env-Datei
	loadEnv()

	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		slog.Error("MISTRAL_API_KEY ist nicht in der .env-Datei gesetzt")
		os.Exit(1)
	}

	chat := []*request.Message{
		{Role: "system", Content: "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."},
	}

	client := &http.Client{}

	newMessages := []*request.Message{
		//{Role: "user", Content: "Welches ist die Haupt Datei in meinem Projekt und welche Datei liegt im Package 'request'"},
	}

	for {

		if len(newMessages) == 0 {
			chat = append(chat, getUserMessage("Was ist dein Begehr"))
		} else {
			chat = append(chat, newMessages...)
			newMessages = []*request.Message{}
		}

		jsonData, err := json.Marshal(&request.ChatCompletion{
			Model:    "devstral-medium-latest",
			Stream:   true,
			Messages: chat,
			Tools:    tools.GetTools(),
		})

		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			panic(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		req.Header.Set("Accept", "text/event-stream")

		slog.Default().Info("send completion request", "url", req.URL.String(), "jsonData", jsonData)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		if resp.StatusCode >= 400 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.Error("Fehler beim Lesen des Response-Body", "error", err)
				resp.Body.Close()
				continue
			}

			resp.Body = io.NopCloser(bytes.NewBuffer(body))

			slog.Error("HTTP-Fehler", "status", resp.StatusCode, "body", string(body))
			resp.Body.Close()
			continue
		}

		reader := bufio.NewReader(resp.Body)

		var completition response.CompletionChunk

		responseMessage := &request.Message{
			Role: "assistant",
		}

		chat = append(chat, responseMessage)

		var builder strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			line = strings.TrimSpace(line)

			// SSE Format: "data: ..."
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)

				if data == "[DONE]" {
					break
				}

				err := json.Unmarshal([]byte(data), &completition)
				if err != nil {
					panic(err)
				}

				first := completition.Choices[0]

				switch first.FinishReason {

				case "tool_calls":

					responseMessage.ToolCalls = append(responseMessage.ToolCalls, first.Delta.ToolCalls...)

					for _, toolCall := range first.Delta.ToolCalls {
						msg, err := call_tool(toolCall)
						if err != nil {
							newMessages = append(newMessages, &request.Message{
								Role:       "tool",
								ToolCallId: toolCall.Id,
								Content:    fmt.Sprintf("Beim Aufruf des Tools ist ein fehler aufgetreten. (error: %v)", err),
							})
						} else {
							newMessages = append(newMessages, &request.Message{
								Role:       "tool",
								ToolCallId: toolCall.Id,
								Content:    string(msg),
							})
						}
					}

				default:
					switch first.Delta.Content.(type) {
					case map[string]any:
						fmt.Printf("MAP %v\n", completition.Choices[0].Delta.Content)
					case []any:
						fmt.Printf("LIST %v\n", completition.Choices[0].Delta.Content)
					case string:
						fmt.Print(completition.Choices[0].Delta.Content.(string))
						builder.WriteString(completition.Choices[0].Delta.Content.(string))

					default:
						fmt.Printf("Unknown type %v\n", completition.Choices[0].Delta.Content)
					}

				}

			}
		}

		if builder.Len() > 0 {
			println()
			responseMessage.Content = builder.String()
		}

		resp.Body.Close()

	}

}

func loadEnv() {
	// Öffne die .env-Datei
	file, err := os.Open(".env")
	if err != nil {
		slog.Error("Fehler beim Öffnen der .env-Datei", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// Lese die Datei zeilenweise
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignoriere leere Zeilen und Kommentare
		if line == "" || line[0] == '#' {
			continue
		}

		// Teile die Zeile in Key und Value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Setze die Umgebungsvariable
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Fehler beim Lesen der .env-Datei", "error", err)
		os.Exit(1)
	}
}

func getUserMessage(s string) *request.Message {
	reader := bufio.NewReader(os.Stdin)
	println(s)
	input, _ := reader.ReadString('\n')
	return &request.Message{
		Role:    "user",
		Content: input,
	}
}

type ReadFileArgs struct {
	Path string `json:"path"`
}

type ReplaceFileContentArgs struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

func call_tool(call response.ToolCallChoice) ([]byte, error) {
	if call.Function.Name == "read_file" {
		var args ReadFileArgs
		json.Unmarshal([]byte(call.Function.Arguments), &args)
		content, err := os.ReadFile(args.Path)
		if err != nil {
			slog.Error("Fehler beim Lesen der Datei", "path", args.Path, "error", err)
			return nil, err
		}
		return content, nil
	} else if call.Function.Name == "replace_file_content" {
		var args ReplaceFileContentArgs
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return nil, err
		} else {
			return replaceFileContent(args)
		}
	} else if call.Function.Name == "list_files" {
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
	} else if call.Function.Name == "write_file" {
		var args WriteFileArgs
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return nil, err
		}
		return writeFile(args)
	} else if call.Function.Name == "git_do" {
		var args GitDoArgs
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return nil, err
		}
		return gitDo(args)
	} else if call.Function.Name == "git_diff" {
		return gitDiff()
	} else {
		return nil, errors.New("tool nicht gefunden")
	}
}

type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func writeFile(args WriteFileArgs) ([]byte, error) {
	// Schreibe den Inhalt in die Datei
	err := os.WriteFile(args.Path, []byte(args.Content), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Datei erfolgreich geschrieben oder erstellt"})
}

type StatusResponse struct {
	Status   string `json:"status"`
	Messsage string `json:"messsage,omitempty"`
	Error    string `json:"error,omitempty"`
}

func replaceFileContent(args ReplaceFileContentArgs) ([]byte, error) {
	// Überprüfe, ob die Datei existiert
	if _, err := os.Stat(args.Path); os.IsNotExist(err) {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Lese den Inhalt der Datei
	content, err := ioutil.ReadFile(args.Path)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Ersetze den Inhalt
	fileContent := string(content)
	var newFileContent string

	if args.ReplaceAll {
		newFileContent = strings.ReplaceAll(fileContent, args.OldContent, args.NewContent)
	} else {
		newFileContent = strings.Replace(fileContent, args.OldContent, args.NewContent, 1)
	}

	// Schreibe den neuen Inhalt zurück in die Datei
	err = os.WriteFile(args.Path, []byte(newFileContent), 0644)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	return json.Marshal(StatusResponse{Status: "OK", Messsage: "Replacement erfolgreich"})
}

type GitDoArgs struct {
	CommitMessage string `json:"commit_message"`
}

func gitDo(args GitDoArgs) ([]byte, error) {
	// Führe git add . aus
	if err := runCommand("git", "add", "."); err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}

	// Führe git commit aus
	if err := runCommand("git", "commit", "-m", args.CommitMessage); err != nil {
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

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Fehler beim Ausführen von %s: %v (stderr: %s)", name, err, stderr.String())
	}

	return nil
}

func gitDiff() ([]byte, error) {
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
