package tools

import (
	"encoding/json"
	"github.com/mwildt/progoter/request"
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

// GetTools liefert die Liste der verfügbaren Tools
func GetTools() []request.Tool {
	return []request.Tool{
		WriteFileTool{}.GetTool(),
		ListFilesTool{}.GetTool(),
		ReadFileTool{}.GetTool(),
		CreateDirTool{}.GetTool(),
		ReplaceFileLinesTool{}.GetTool(),
		EditFileTool{}.GetTool(),
		//ReplaceFileContentTool{}.GetTool(),
		GitDoTool{}.GetTool(),
		GitDiffTool{}.GetTool(),
		StopProcessTool{}.GetTool(),
		GolangTool{}.GetTool(),
	}
}
