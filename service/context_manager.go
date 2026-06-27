package service

import (
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

// CreateContext erstellt einen neuen Chat-Kontext mit der gegebenen ID.
func (cm *ContextManager) CreateContext(id string) *ChatContext {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	context := NewChatContext()
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
	cm.contexts[id] = NewChatContext()
}

// DeleteContext löscht den Chat-Kontext für die gegebene ID.
func (cm *ContextManager) DeleteContext(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.contexts, id)
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
