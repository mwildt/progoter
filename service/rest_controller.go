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
	chatService  *ChatService
	chatContext  *ChatContext
	mu           sync.Mutex
	isProcessing bool
}

func NewRESTController(apiKey string) *RESTController {
	return &RESTController{
		chatService:  NewChatService(apiKey),
		chatContext:  NewChatContext(),
		isProcessing: false,
	}
}

// StartProcessing setzt den Processing-Status auf true.
func (rc *RESTController) StartProcessing() bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.isProcessing {
		return false
	}
	rc.isProcessing = true
	return true
}

// StopProcessing setzt den Processing-Status auf false.
func (rc *RESTController) StopProcessing() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.isProcessing = false
}

// ClearContextHandler setzt den Chat-Kontext zurück.
func (rc *RESTController) ClearContextHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.StartProcessing() {
		http.Error(w, "Eine Aktion läuft bereits", http.StatusConflict)
		return
	}
	defer rc.StopProcessing()

	err := rc.ClearContext()
	if err != nil {
		slog.Error("Fehler beim Zurücksetzen des Chatverlaufs", "error", err)
		http.Error(w, "Fehler beim Zurücksetzen des Chatverlaufs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chat-Kontext wurde erfolgreich zurückgesetzt.")
}

// DumpContextHandler speichert den aktuellen Chat-Kontext als JSON-Datei.
func (rc *RESTController) DumpContextHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.StartProcessing() {
		http.Error(w, "Eine Aktion läuft bereits", http.StatusConflict)
		return
	}
	defer rc.StopProcessing()

	err := rc.DumpContext()
	if err != nil {
		slog.Error("Fehler beim Schreiben des Chatverlaufs", "error", err)
		http.Error(w, "Fehler beim Schreiben des Chatverlaufs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chat-Kontext wurde erfolgreich gespeichert.")
}

// CompactChatHandler komprimiert den Chatverlauf.
func (rc *RESTController) CompactChatHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.StartProcessing() {
		http.Error(w, "Eine Aktion läuft bereits", http.StatusConflict)
		return
	}
	defer rc.StopProcessing()

	err := rc.CompactChat()
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

// MessageHandler sendet eine Nachricht und liefert einen SSE-Stream mit Events.
func (rc *RESTController) MessageHandler(w http.ResponseWriter, r *http.Request) {

	go func() {
		<-r.Context().Done()
		slog.Info("request abgebrochen", "error", r.Context().Err())
	}()

	if !rc.StartProcessing() {
		http.Error(w, "Eine Aktion läuft bereits", http.StatusConflict)
		return
	}
	defer rc.StopProcessing()

	var messageRequest PostMessageRequestDTO
	err := json.NewDecoder(r.Body).Decode(&messageRequest)
	if err != nil {
		http.Error(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	// Setze den Content-Type-Header für SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Füge die Nachricht zum Chat-Kontext hinzu
	rc.chatContext.AddMessage(&request.Message{
		Role:    "user",
		Content: messageRequest.Message,
	})

	messageChan := make(chan *request.Message)

	// Erstelle einen Kanal für die SSE-Ereignisse
	sseChan := make(chan string)
	go func() {
		defer close(sseChan)
		for msg := range messageChan {
			if len(msg.Content) == 0 {
				continue
			}
			data, _ := json.Marshal(msg)
			sseChan <- fmt.Sprintf("data: %s\n\n", string(data))
		}
	}()

	// Verarbeite die Nachricht
	var errError error
	go func() {
		defer func() {
			if errError != nil {
				slog.Error("Fehler beim Verarbeiten der Chat-Vervollständigung", "error", errError)
			}
		}()
		_, errError = rc.chatService.CompleteContext(r.Context(), rc.chatContext, messageChan)
	}()

	// Sende SSE-Ereignisse an den Client
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming nicht unterstützt", http.StatusInternalServerError)
		return
	}

	for event := range sseChan {
		fmt.Fprintf(w, event)
		flusher.Flush()
	}

	if errError != nil {
		http.Error(w, "Fehler beim Verarbeiten der Chat-Vervollständigung", http.StatusInternalServerError)
	}
}

// CompactChat komprimiert den Chatverlauf.
func (rc *RESTController) CompactChat() error {
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

	rc.chatContext.AddMessage(summarizeMessage)
	var messageChan chan *request.Message = nil

	compactedContext, err := rc.chatService.CompleteContext(context.Background(), rc.chatContext, messageChan)
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

	rc.chatContext = newContext
	return nil
}

// DumpContext speichert den aktuellen Chat-Kontext als JSON-Datei.
func (rc *RESTController) DumpContext() error {
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
	err = encoder.Encode(rc.chatContext)
	if err != nil {
		return fmt.Errorf("Fehler beim Schreiben des JSON: %v", err)
	}

	return nil
}

// ClearContext setzt den Chat-Kontext zurück.
func (rc *RESTController) ClearContext() error {
	systemPrompt, err := readSystemPrompt()
	if err != nil {
		systemPrompt = "Du bist ein hilfreicher Agent bei der Programmierung von golang apps."
	}

	newContext := &ChatContext{
		Messages: []*request.Message{
			{Role: "system", Content: systemPrompt},
		},
	}

	rc.chatContext = newContext
	return nil
}

// GetContextHandler gibt den aktuellen Chat-Kontext als JSON zurück.
func (rc *RESTController) GetContextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(rc.chatContext)
	if err != nil {
		slog.Error("Fehler beim Kodieren des Chat-Kontexts", "error", err)
		http.Error(w, "Fehler beim Kodieren des Chat-Kontexts", http.StatusInternalServerError)
		return
	}
}

// SetupRoutes richtet die REST-Routen ein.
func (rc *RESTController) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /chat/default/clear", rc.ClearContextHandler)
	mux.HandleFunc("GET /chat/default/dump", rc.DumpContextHandler)
	mux.HandleFunc("POST /chat/default/compact", rc.CompactChatHandler)
	mux.HandleFunc("POST /chat/default/message", rc.MessageHandler)
	mux.HandleFunc("GET /chat/default/context", rc.GetContextHandler)
}
