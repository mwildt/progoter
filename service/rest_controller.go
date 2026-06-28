package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type RESTController struct {
	chatService    *ChatService
	contextManager *ContextManager
	mu             sync.Mutex
}

func NewRESTController(apiKey string) *RESTController {
	return &RESTController{
		chatService:    NewChatService(apiKey),
		contextManager: NewContextManager(),
	}
}

// ClearContextHandler setzt den Chat-Kontext für die gegebene ID zurück.
func (rc *RESTController) ClearContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	chatContext, exists := rc.contextManager.GetContext(id)
	if !exists {
		http.Error(w, "Chat-Kontext nicht gefunden", http.StatusNotFound)
		return
	}

	err := chatContext.ClearMessages()
	if err != nil {
		slog.Error("Fehler beim Zurücksetzen des Chatverlaufs", "error", err)
		http.Error(w, "Fehler beim Zurücksetzen des Chatverlaufs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chat-Kontext wurde erfolgreich zurückgesetzt.")
}

// DumpContextHandler speichert den Chat-Kontext für die gegebene ID als JSON-Datei.
func (rc *RESTController) DumpContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	chatContext, exists := rc.contextManager.GetContext(id)
	if !exists {
		http.Error(w, "Chat-Kontext nicht gefunden", http.StatusNotFound)
		return
	}

	err := rc.dumpContext(chatContext)
	if err != nil {
		slog.Error("Fehler beim Schreiben des Chatverlaufs", "error", err)
		http.Error(w, "Fehler beim Schreiben des Chatverlaufs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chat-Kontext wurde erfolgreich gespeichert.")
}

// CompactChatHandler komprimiert den Chatverlauf für die gegebene ID.
func (rc *RESTController) CompactChatHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	chatContext, exists := rc.contextManager.GetContext(id)
	if !exists {
		http.Error(w, "Chat-Kontext nicht gefunden", http.StatusNotFound)
		return
	}

	err := rc.compactChat(chatContext)
	if err != nil {
		slog.Error("Fehler beim Komprimieren des Chatverlaufs", "error", err)
		http.Error(w, "Fehler beim Komprimieren des Chatverlaufs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chatverlauf wurde erfolgreich komprimiert.")
}

type PostMessageRequestDTO struct {
	Message string `json:"message"`
}

// PostMessageHandler sendet eine Nachricht und liefert einen SSE-Stream mit Events.
func (rc *RESTController) PostMessageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	chatContext, exists := rc.contextManager.GetContext(id)
	if !exists {
		chatContext = rc.contextManager.CreateContext(id)
	}

	var messageRequest PostMessageRequestDTO
	err := json.NewDecoder(r.Body).Decode(&messageRequest)
	if err != nil {
		http.Error(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	// Füge die Nachricht zum Chat-Kontext hinzu
	message := &request.Message{
		Role:    "user",
		Content: messageRequest.Message,
	}
	chatContext.AddMessage(message)
	chatContext.Broadcast(message)

	if chatContext.Complete(rc.chatService); err != nil {
		http.Error(w, "Fehler beim Verarbeiten der Chat-Vervollständigung", http.StatusInternalServerError)
	}

}

// compactChat komprimiert den gegebenen Chat-Kontext.
func (rc *RESTController) compactChat(chatContext *ChatContext) error {
	summarizeMessage := &request.Message{
		Role: "user",
		Content: `
Fasse den bisherigen Chatverlauf zusammen. Ziel ist es, **alle fachlichen Informationen, Entscheidungen, Daten, Code-Snippets und Kontext** zu erhalten. Ignoriere dabei:
- Smalltalk
- Bestätigungen ("Ja", "Okay", "Verstanden")
- Wiederholungen
- Off-Topic-Diskussionen

**Regeln:**
0. Ignoriere System nachrichten .
1. Behalte **technische Details, Anforderungen, Lösungsansätze und offene Fragen** bei.
2. Strukturiere die Zusammenfassung nach Themen/Abschnitten (z. B. "Anforderungen", "Technische Umsetzung", "Offene Punkte").
3. Verwende **die gleiche Terminologie** wie im Original.
4. Falls Code oder Daten im Verlauf vorkommen: **Füge sie unverändert ein** (keine Paraphrasierung).
5. Maximal 50 % der ursprünglichen Token-Anzahl.

**Ausgabeformat:**
- Knappe, klare Sätze.
- Keine Erklärungen, warum etwas zusammengefasst wurde.
- Keine Einleitungen wie "Hier ist die Zusammenfassung:".
`,
	}

	chatContext.AddMessage(summarizeMessage)
	var messageChan chan *request.Message = nil

	compactedContext, err := rc.chatService.CompleteContext(context.Background(), chatContext, messageChan)
	if err != nil {
		return fmt.Errorf("Fehler beim Komprimieren des Chatverlaufs: %v", err)
	}

	var summary string
	messages := compactedContext.GetMessages()
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			summary = messages[i].Content
			break
		}
	}

	systemPrompt, err := readSystemPrompt()
	if err != nil {
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	newContext := &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "assistant", Content: summary},
		},
	}

	chatContext.Messages = newContext.Messages
	return nil
}

// dumpContext speichert den gegebenen Chat-Kontext als JSON-Datei.
func (rc *RESTController) dumpContext(chatContext *ChatContext) error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("dumps/context.%s.json", timestamp)

	err := os.MkdirAll("dumps", os.ModePerm)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen des Verzeichnisses: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen der Datei: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(chatContext)
	if err != nil {
		return fmt.Errorf("Fehler beim Schreiben des JSON: %v", err)
	}

	return nil
}

// ClearContext setzt den Chat-Kontext zurück.
func (rc *RESTController) ClearContext(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	systemPrompt, err := readSystemPrompt()
	if err != nil {
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	newContext := &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: systemPrompt},
		},
	}

	rc.contextManager.SetContext(id, newContext)
	return nil
}

// GetContextHandler gibt den Chat-Kontext für die gegebene ID als Stream zurück.
func (rc *RESTController) GetContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	chatContext, exists := rc.contextManager.GetContext(id)

	if !exists {
		chatContext = rc.contextManager.CreateContext(id)
	}

	// Setze den Content-Type-Header für SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming nicht unterstützt", http.StatusInternalServerError)
		return
	}

	// Abonniere den Chat-Kontext
	sub := chatContext.Stream()
	for msg := range sub {
		data, err := json.Marshal(msg)
		if err != nil {
			slog.Error("Fehler beim Kodieren der Nachricht", "error", err)
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}

}

// SetupRoutes richtet die REST-Routen ein.
func (rc *RESTController) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /chat/{id}/clear", rc.ClearContextHandler)
	mux.HandleFunc("GET /chat/{id}/dump", rc.DumpContextHandler)
	mux.HandleFunc("POST /chat/{id}/compact", rc.CompactChatHandler)
	mux.HandleFunc("POST /chat/{id}/message", rc.PostMessageHandler)
	mux.HandleFunc("GET /chat/{id}/context", rc.GetContextHandler)
}
