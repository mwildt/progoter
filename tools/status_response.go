package tools

// StatusResponse enthält den Status und eine optionale Nachricht oder Fehler
type StatusResponse struct {
	Status   string `json:"status"`
	Messsage string `json:"messsage,omitempty"`
	Error    string `json:"error,omitempty"`
}
