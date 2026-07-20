package tools

import (
	"encoding/json"
	"strings"
)

type ToolError string

func (e ToolError) Error() string {
	return string(e)
}

const ParseError = ToolError("Fehler beim Parsen der Argumente")

func errorResponse(message string, err error) ([]byte, error) {
	status := StatusResponse{
		Status:  "ERROR",
		Message: message,
		Error:   err.Error(),
	}
	jsonData, jsonErr := json.Marshal(status)
	if jsonErr != nil {
		return jsonData, jsonErr
	} else {
		return jsonData, err
	}
}

func successResponse(message string) ([]byte, error) {
	status := StatusResponse{Status: "OK", Message: message}
	return json.Marshal(status)
}

type FileExclusion string

func (e FileExclusion) Match(path string) bool {
	//TODO: Hier wäre mal eine richtiger parser einzusetzen.
	return strings.HasPrefix(path, string(e))
}

type FileExclusions []FileExclusion

func (e FileExclusions) Match(path string) bool {
	for _, ex := range e {
		if ex.Match(path) {
			return true
		}
	}
	return false
}
