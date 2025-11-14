// Package models - Ollama model adapter and internal factory
package models

import (
	"context"
	"fmt"
	"iter"
	"net/url"

	"github.com/ollama/ollama/api"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// OllamaModelAdapter implements the model.LLM interface for Ollama models
type OllamaModelAdapter struct {
	client    *api.Client
	modelName string
}

// createOllamaModelInternal creates a model using the Ollama API backend (internal implementation)
func createOllamaModelInternal(ctx context.Context, cfg OllamaConfig) (model.LLM, error) {
	if cfg.ModelName == "" {
		return nil, fmt.Errorf("Ollama model name is required")
	}

	// Use provided host or let the client use the default from environment
	var client *api.Client
	if cfg.Host != "" {
		// Parse and create a custom client with the provided host
		baseURL, err := url.Parse(cfg.Host)
		if err != nil {
			return nil, fmt.Errorf("invalid Ollama host URL: %w", err)
		}
		client = api.NewClient(baseURL, nil)
	} else {
		// Use default client from environment (OLLAMA_HOST env var or default local endpoint)
		var err error
		client, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client: %w", err)
		}
	}

	// Test connectivity to the Ollama server
	if err := client.Heartbeat(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect to Ollama server: %w", err)
	}

	return &OllamaModelAdapter{
		client:    client,
		modelName: cfg.ModelName,
	}, nil
}

// Name returns the model name
func (a *OllamaModelAdapter) Name() string {
	return a.modelName
}

// GenerateContent implements the model.LLM interface
// It handles streaming and non-streaming requests to Ollama's /api/chat endpoint
func (a *OllamaModelAdapter) GenerateContent(
	ctx context.Context,
	req *model.LLMRequest,
	stream bool,
) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// Convert genai.Content to Ollama chat messages
		messages, err := convertToOllamaMessages(req.Contents)
		if err != nil {
			yield(nil, fmt.Errorf("failed to convert contents to Ollama messages: %w", err))
			return
		}

		// Extract model name from request if not set
		modelName := a.modelName
		if req.Model != "" {
			modelName = req.Model
		}

		// Prepare Ollama chat request
		ollamaReq := &api.ChatRequest{
			Model:    modelName,
			Messages: messages,
			Stream:   &stream,
		}

		// Apply config if provided
		if req.Config != nil {
			ollamaReq.Options = make(map[string]any)

			if req.Config.Temperature != nil {
				ollamaReq.Options["temperature"] = *req.Config.Temperature
			}

			if req.Config.TopP != nil {
				ollamaReq.Options["top_p"] = *req.Config.TopP
			}

			if req.Config.MaxOutputTokens > 0 {
				ollamaReq.Options["num_predict"] = req.Config.MaxOutputTokens
			}

			// Handle tools if provided
			if len(req.Config.Tools) > 0 {
				tools, err := convertToOllamaTools(req.Config.Tools)
				if err != nil {
					yield(nil, fmt.Errorf("failed to convert tools: %w", err))
					return
				}
				ollamaReq.Tools = tools
			}
		}

		// Accumulate the full response for non-streaming mode
		var fullContent string
		var fullResponse *api.ChatResponse

		// Use the Chat API with streaming
		err = a.client.Chat(ctx, ollamaReq, func(resp api.ChatResponse) error {
			// For streaming responses, convert each chunk
			if stream {
				modelResp := convertOllamaChatResponseToGenAI(resp, true)
				if !yield(modelResp, nil) {
					return context.Canceled
				}
			} else {
				// Accumulate content for non-streaming
				fullContent += resp.Message.Content
				fullResponse = &resp
			}
			return nil
		})

		if err != nil {
			yield(nil, fmt.Errorf("Ollama chat request failed: %w", err))
			return
		}

		// For non-streaming mode, yield the accumulated response
		if !stream && fullResponse != nil {
			fullResponse.Message.Content = fullContent
			modelResp := convertOllamaChatResponseToGenAI(*fullResponse, false)
			yield(modelResp, nil)
		}
	}
}

// convertToOllamaMessages converts genai.Content to Ollama API messages
func convertToOllamaMessages(contents []*genai.Content) ([]api.Message, error) {
	var messages []api.Message

	for _, content := range contents {
		if content == nil {
			continue
		}

		// Get role name from content (should be "user", "assistant", "system", etc.)
		role := "user" // default
		if content.Role != "" {
			role = content.Role
		}

		// Convert parts to a single string message
		var contentText string
		for _, part := range content.Parts {
			if part == nil {
				continue
			}

			// Handle text content
			if part.Text != "" {
				contentText += part.Text
			}
			// TODO: Handle images/blobs if needed for multimodal models
		}

		messages = append(messages, api.Message{
			Role:    role,
			Content: contentText,
		})
	}

	return messages, nil
}

// convertToOllamaTools converts genai.Tool to Ollama API tools
func convertToOllamaTools(tools []*genai.Tool) ([]api.Tool, error) {
	var ollamaTools []api.Tool

	for _, tool := range tools {
		if tool.FunctionDeclarations == nil {
			continue
		}

		for _, funcDecl := range tool.FunctionDeclarations {
			params := api.ToolFunctionParameters{
				Type:       "object",
				Properties: make(map[string]api.ToolProperty),
			}

			// Convert parameters if available
			if funcDecl.Parameters != nil {
				for propName, prop := range funcDecl.Parameters.Properties {
					params.Properties[propName] = convertSchemaPropertyToTool(prop)
				}
				params.Required = funcDecl.Parameters.Required
			}

			ollamaTools = append(ollamaTools, api.Tool{
				Type: "function",
				Function: api.ToolFunction{
					Name:        funcDecl.Name,
					Description: funcDecl.Description,
					Parameters:  params,
				},
			})
		}
	}

	return ollamaTools, nil
}

// convertSchemaPropertyToTool converts a genai schema property to an Ollama ToolProperty
func convertSchemaPropertyToTool(prop *genai.Schema) api.ToolProperty {
	toolProp := api.ToolProperty{
		Description: prop.Description,
	}

	if prop.Type != "" {
		toolProp.Type = api.PropertyType{string(prop.Type)}
	}

	// Convert enum values - Ollama expects []any
	if len(prop.Enum) > 0 {
		enumValues := make([]any, len(prop.Enum))
		for i, v := range prop.Enum {
			enumValues[i] = v
		}
		toolProp.Enum = enumValues
	}

	return toolProp
}

// convertOllamaChatResponseToGenAI converts Ollama ChatResponse to model.LLMResponse
func convertOllamaChatResponseToGenAI(resp api.ChatResponse, isStreaming bool) *model.LLMResponse {
	// Create the response content with text part
	parts := []*genai.Part{
		{
			Text: resp.Message.Content,
		},
	}

	// Convert tool calls if present
	if len(resp.Message.ToolCalls) > 0 {
		for _, toolCall := range resp.Message.ToolCalls {
			// Convert arguments to proper format
			args := make(map[string]any)
			if len(toolCall.Function.Arguments) > 0 {
				for k, v := range toolCall.Function.Arguments {
					args[k] = v
				}
			}

			parts = append(parts, &genai.Part{
				FunctionCall: &genai.FunctionCall{
					ID:   toolCall.ID,
					Name: toolCall.Function.Name,
					Args: args,
				},
			})
		}
	}

	// Create the response content
	responseContent := &genai.Content{
		Role:  "model",
		Parts: parts,
	}

	// Create the LLMResponse
	modelResp := &model.LLMResponse{
		Content:      responseContent,
		Partial:      isStreaming && resp.Done == false,
		TurnComplete: resp.Done,
		FinishReason: stringToFinishReason(resp.DoneReason),
	}

	return modelResp
}

// stringToFinishReason converts Ollama done_reason string to genai.FinishReason
func stringToFinishReason(reason string) genai.FinishReason {
	switch reason {
	case "stop":
		return genai.FinishReasonStop
	case "length":
		return genai.FinishReasonMaxTokens
	case "tool_calls":
		// Map tool_calls to stop (there's no dedicated FinishReasonToolCalls in genai)
		return genai.FinishReasonStop
	default:
		return genai.FinishReasonOther
	}
}
