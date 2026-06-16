package main

import (
	"bufio"
	"github.com/mwildt/progoter/service"
	"log/slog"
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

	cliController := service.NewCLIController(apiKey)
	cliController.StartChat()
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
