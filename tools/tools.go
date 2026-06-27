package tools

import "github.com/mwildt/progoter/request"

// GetTools liefert die Liste der verfügbaren Tools
func GetTools() []request.Tool {
	return []request.Tool{
		WriteFileTool{}.GetTool(),
		ListFilesTool{}.GetTool(),
		ReadFileTool{}.GetTool(),
		CreateDirTool{}.GetTool(),
		ReplaceFileLinesTool{}.GetTool(),
		ReplaceFileContentTool{}.GetTool(),
		GitDoTool{}.GetTool(),
		GitDiffTool{}.GetTool(),
		StopProcessTool{}.GetTool(),
	}
}
