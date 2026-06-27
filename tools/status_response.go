package tools

// StatusResponse enthält den Status und eine optionale Nachricht oder Fehler
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
