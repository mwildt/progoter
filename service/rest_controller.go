package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mwildt/progoter/request"
	"log/slog"
	"net/http"
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

// DumpContextHandler speichert den Chat-Kontext für die gegebene ID als JSON-Datei.
func (rc *RESTController) DumpContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	if chatContext, exists := rc.contextManager.GetContext(id); !exists {
		http.NotFound(w, r)
	} else {
		chatContext.Dump()
	}
}

// PostCompactChatHandler komprimiert den Chatverlauf für die gegebene ID.
func (rc *RESTController) PostCompactChatHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	if chatContext, exists := rc.contextManager.GetContext(id); !exists {
		http.NotFound(w, r)
	} else {
		chatContext.Compcat(rc.chatService)
	}
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

func (rc *RESTController) PostClearContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	if chatContext, exists := rc.contextManager.GetContext(id); !exists {
		http.NotFound(w, r)
	} else {
		chatContext.Clear()
	}
}

func randomString(length int) (string, error) {
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func (rc *RESTController) GetContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	trace, _ := randomString(9)

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
	for event := range sub {
		switch event.(type) {
		case *request.Message:
			data, _ := json.Marshal(event)
			//slog.Default().With("logger", "RESTController", "trace", trace).
			fmt.Fprintf(w, "id: %d\n", time.Now().UnixMicro())
			fmt.Fprintf(w, "event: %s\n", "chat-message")
			fmt.Fprintf(w, "data: %s\n", string(data))
			fmt.Fprintf(w, "\n")
			flusher.Flush()

		case StateEvent:
			slog.Default().With("logger", "RESTController", "trace", trace).
				Info("send message", "type", "state-change")
			fmt.Fprintf(w, "id: %d\n", time.Now().UnixMicro())
			fmt.Fprintf(w, "event: %s\n", "state-change")
			if StateIdle == event {
				fmt.Fprintf(w, "data: %s\n", "idle")
			} else if StateProcessing == event {
				fmt.Fprintf(w, "data: %s\n", "processing")
			} else {
				fmt.Fprintf(w, "data: %s\n", "unknown")
			}
			fmt.Fprintf(w, "\n")
			flusher.Flush()
		}

	}

}

// CancelContextHandler stoppt den aktuellen Chat-Kontext für die gegebene ID.
func (rc *RESTController) CancelContextHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = "default"
	}

	if chatContext, exists := rc.contextManager.GetContext(id); !exists {
		http.NotFound(w, r)
	} else {
		chatContext.Cancel()
	}
}

// SetupRoutes richtet die REST-Routen ein.
func (rc *RESTController) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /chat/{id}/clear", rc.PostClearContextHandler)
	mux.HandleFunc("GET /chat/{id}/dump", rc.DumpContextHandler)
	mux.HandleFunc("POST /chat/{id}/compact", rc.PostCompactChatHandler)
	mux.HandleFunc("POST /chat/{id}/message", rc.PostMessageHandler)
	mux.HandleFunc("GET /chat/{id}/context", rc.GetContextHandler)
	mux.HandleFunc("POST /chat/{id}/cancel", rc.CancelContextHandler)
}
