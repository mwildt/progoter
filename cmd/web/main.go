package main

import (
	"bufio"
	"errors"
	"github.com/mwildt/progoter/chat"
	"github.com/mwildt/progoter/chatapi"
	"github.com/mwildt/progoter/tools"
	"log"
	"log/slog"
	"net/http"
	"os"
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

	workspaceDir := os.Getenv("WORKSPACE_DIR")
	if workspaceDir == "" {
		workspaceDir = "./"
	}

	if info, err := os.Stat(workspaceDir); errors.Is(err, os.ErrNotExist) || !info.IsDir() {
		log.Fatalf("workspaceDir %s is not a directory", workspaceDir)
	}

	slog.Default().Info("starting application", "workspaceDir", workspaceDir)

	apiService := chatapi.NewService(apiKey)

	toolService := tools.NewService(
		tools.AllTools(),
		tools.WorkspaceDir(workspaceDir),
	)

	// wir haben hier eine zyklische Abhängikeit zwischen dem subagent tool und dem tool service. Das keiegen wir auch so einfach nicht weg
	toolService.Configure(tools.ToolConfig(tools.NewSubagentTool(apiService, toolService)))

	chatService := chat.NewChatService(toolService, apiService)
	restController := chat.NewRESTController(chatService, chat.NewContextManager())

	mux := http.NewServeMux()
	restController.SetupRoutes(mux)

	fs := http.FileServer(http.Dir("./web/resources"))
	mux.Handle("/", fs)

	slog.Info("Server wird gestartet auf Port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("Fehler beim Starten des Servers", "error", err)
		os.Exit(1)
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
