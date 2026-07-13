package tools

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	workspaceDir string
	tools        map[string]ToolHandler
}

type ServiceConfig func(*Service)

func ToolConfig(tools ...ToolHandler) ServiceConfig {
	return func(service *Service) {
		for _, tool := range tools {
			meta := tool.GetTool()
			service.tools[meta.Function.Name] = tool
		}
	}
}

func WorkspaceDir(workstaeDir string) ServiceConfig {
	return func(service *Service) {
		service.workspaceDir = workstaeDir
	}
}

func NewService(configs ...ServiceConfig) *Service {
	service := &Service{
		workspaceDir: "./",
		tools:        make(map[string]ToolHandler),
	}
	for _, config := range configs {
		config(service)
	}
	return service
}

func AllTools() ServiceConfig {
	return ToolConfig(
		&ReadFileTool{},
		&EditFileTool{},
		&ListFilesTool{},
		&WriteFileTool{},
		&GitDoTool{},
		&GitDiffTool{},
		&CreateDirTool{},
		//&StopProcessTool{},
		&CheckTool{},
		&ReplaceFileLinesTool{},
		&GolangTool{},
		&SearchInFilesTool{Exclusions: FileExclusions{".idea/", ".git/"}},
	)
}

func (s Service) GetTools(_ context.Context) (result []ToolDefinition) {
	for _, tool := range s.tools {
		result = append(result, tool.GetTool())
	}
	return result
}

func (s Service) CallFunction(ctx context.Context, baseDir string, name string, args string) ([]byte, error) {
	if tool, exists := s.tools[name]; exists {
		return tool.Execute(baseDir, args)
	} else {
		return errorResponse(fmt.Sprintf("tool '%s' not found", name), errors.New("tool not found"))
	}
}
