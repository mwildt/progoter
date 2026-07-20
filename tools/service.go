package tools

import "C"
import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

func (s *Service) GetTools(_ context.Context) (result []ToolDefinition) {
	for _, tool := range s.tools {
		result = append(result, tool.GetTool())
	}
	return result
}

type ToolFilter func(ToolHandler) bool

func HasType[T ToolHandler]() ToolFilter {
	return func(tool ToolHandler) bool {
		_, ok := tool.(T)
		return ok
	}
}

func Not(filter ToolFilter) ToolFilter {
	return func(tool ToolHandler) bool {
		return !filter(tool)
	}
}

func And(filters ...ToolFilter) ToolFilter {
	return func(tool ToolHandler) bool {
		for _, filter := range filters {
			if !filter(tool) {
				return false
			}
		}
		return true
	}
}

func (s *Service) FilterTools(_ context.Context, filter ToolFilter) (result []ToolDefinition) {
	for _, tool := range s.tools {
		if filter(tool) {
			result = append(result, tool.GetTool())
		}
	}
	return result
}

func (s *Service) GetOperatingDir(baseDir string) (string, error) {

	absWorkspaceDir, err := filepath.Abs(s.workspaceDir)
	if err != nil {
		return "", err
	}

	operatingDir, err := filepath.Abs(filepath.Join(absWorkspaceDir, baseDir))
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(filepath.Clean(operatingDir), filepath.Clean(absWorkspaceDir)) {
		return "", errors.New("Illegal Dir")
	}

	return operatingDir, nil
}

func (s *Service) CallFunction(ctx context.Context, baseDir string, name string, args string) ([]byte, error) {

	operatingDir, err := s.GetOperatingDir(baseDir)
	if err != nil {
		return errorResponse(fmt.Sprintf("fehler beim Aufruf von '%s', name"), err)
	}

	if tool, exists := s.tools[name]; exists {
		return tool.Execute(operatingDir, args)
	} else {
		return errorResponse(fmt.Sprintf("tool '%s' not found", name), errors.New("tool not found"))
	}
}

func (s *Service) Configure(config ServiceConfig) {
	config(s)
}
