package tools

import (
	"context"
	"encoding/json"
	"github.com/mwildt/progoter/chatapi"
)

type SubagentTool struct {
	api     *chatapi.Service
	service *Service
}

func NewSubagentTool(api *chatapi.Service, service *Service) *SubagentTool {
	return &SubagentTool{api, service}
}

func (t SubagentTool) GetTool() ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: chatapi.ToolFunction{
			Name:        "subagent",
			Description: "startet einen Subagent",
			Parameters: chatapi.FunctionParams{
				Type: "object",
				Properties: map[string]chatapi.ArgumentProperty{
					"role": {
						Type:        "string",
						Name:        "role",
						Description: "Rolle des Subagents. Falls ein spezifischer Systemprompt für die Rolle definiert ist, wird diese verwendet",
					},
					"message": {
						Type:        "string",
						Name:        "role",
						Description: "Nachricht an den Subagenten. Hier kann die Aufgabe beschrieben werden (prompt)",
					},
				},
				Required: []string{},
			},
		},
	}
}

type SubagentArgs struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

func (t SubagentTool) Execute(basePath string, args string) ([]byte, error) {
	var subagentArgs SubagentArgs
	err := json.Unmarshal([]byte(args), &subagentArgs)
	if err != nil {
		status, _ := json.Marshal(StatusResponse{Status: "ERROR", Error: err.Error()})
		return status, err
	}
	var response chatapi.Message

	ctx := context.Background()

	var tools []chatapi.Tool
	for _, t := range t.service.FilterTools(ctx, Not(HasType[SubagentTool]())) {
		tools = append(tools, chatapi.Tool(t))
	}
	_, err = t.api.Complete(ctx, chatapi.ChatCompletionRequest{
		Model:  "devstral-medium-latest",
		Stream: true,
		Messages: []*chatapi.ChatCompletionMessage{{
			Role:    "system",
			Content: "Du bist ein hilfreicher subagent",
		}, {
			Role:    "user",
			Content: subagentArgs.Message,
		}},
		Tools: tools,
	}, &response)
	if err != nil {
		return errorResponse("Fehler beim aufruf des servie", err)
	}
	message := response.Content
	return successResponse(message)
}
