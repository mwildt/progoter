package chat

import (
	"log/slog"
	"sync"
)

// ContextManager verwaltet mehrere Chat-Kontexte.
type ContextManager struct {
	contexts map[string]*ChatContext
	mu       sync.Mutex
}

// NewContextManager erstellt einen neuen ContextManager.
func NewContextManager() *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*ChatContext),
	}
}

// CreateContext erstellt einen neuen Chat-Kontext mit der gegebenen ID und einem Arbeitsverzeichnis.
func (cm *ContextManager) CreateContext(id string, basePath string) *ChatContext {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	slog.Default().Info("CreateContext", "id", id, "basePath", basePath)
	context := NewChatContext(basePath)
	cm.contexts[id] = context
	return context
}

// GetContext gibt den Chat-Kontext für die gegebene ID zurück.
func (cm *ContextManager) GetContext(id string) (*ChatContext, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	context, exists := cm.contexts[id]
	return context, exists
}

func (cm *ContextManager) SetContext(id string, cc *ChatContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.contexts[id] = cc
}

// DeleteContext löscht den Chat-Kontext für die gegebene ID.
func (cm *ContextManager) DeleteContext(id string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	_, exists := cm.contexts[id]
	if exists {
		delete(cm.contexts, id)
	}
	return exists
}

// ListContexts gibt eine Liste aller verfügbaren Kontext-IDs zurück.
func (cm *ContextManager) ListContexts() []string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ids := make([]string, 0, len(cm.contexts))
	for id := range cm.contexts {
		ids = append(ids, id)
	}
	return ids
}
