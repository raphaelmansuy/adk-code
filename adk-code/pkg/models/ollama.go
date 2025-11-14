// Package models - Ollama model adapter and internal factory
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// ollamaToolCallAccumulator accumulates tool call data across streaming chunks
type ollamaToolCallAccumulator struct {
	id        string
	name      string
	arguments map[string]any
}

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

	// Map internal model IDs to actual Ollama model names
	actualModelName := mapOllamaModelName(cfg.ModelName)

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
		modelName: actualModelName,
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

		// Track accumulated tool calls across streaming chunks (tool calls only appear in final chunk)
		toolCallsAccum := make(map[int]*ollamaToolCallAccumulator)

		// Use the Chat API with streaming
		err = a.client.Chat(ctx, ollamaReq, func(resp api.ChatResponse) error {
			// For streaming responses, accumulate and convert
			if stream {
				// Accumulate content text
				fullContent += resp.Message.Content

				// Accumulate tool calls (they only appear in the final chunk when done=true)
				if len(resp.Message.ToolCalls) > 0 {
					for _, toolCall := range resp.Message.ToolCalls {
						// Use function index as key if available, otherwise just accumulate
						idx := 0
						if toolCall.Function.Index > 0 {
							idx = int(toolCall.Function.Index)
						}

						accum, exists := toolCallsAccum[idx]
						if !exists {
							accum = &ollamaToolCallAccumulator{}
							toolCallsAccum[idx] = accum
						}

						// Update fields (in final chunk, these will be complete)
						if toolCall.ID != "" {
							accum.id = toolCall.ID
						}
						if toolCall.Function.Name != "" {
							accum.name = toolCall.Function.Name
						}
						// Arguments come as a map directly from Ollama
						if toolCall.Function.Arguments != nil {
							accum.arguments = toolCall.Function.Arguments
						}
					}
				}

				// Build response with current state
				parts := []*genai.Part{}

				// Add text content if present
				if fullContent != "" {
					parts = append(parts, &genai.Part{
						Text: fullContent,
					})
				}

				// Add accumulated tool calls if any
				if len(toolCallsAccum) > 0 {
					for idx := 0; idx < len(toolCallsAccum); idx++ {
						if accum, exists := toolCallsAccum[idx]; exists {
							args := accum.arguments
							if args == nil {
								args = make(map[string]any)
							}

							parts = append(parts, &genai.Part{
								FunctionCall: &genai.FunctionCall{
									ID:   accum.id,
									Name: accum.name,
									Args: args,
								},
							})
						}
					}
				}

				// Only yield if we have something to say
				if len(parts) > 0 {
					modelResp := &model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: parts,
						},
						Partial:      !resp.Done,
						TurnComplete: resp.Done,
						FinishReason: stringToFinishReason(resp.DoneReason),
					}
					if !yield(modelResp, nil) {
						return context.Canceled
					}
					// Clear accumulated content and tool calls after yielding
					// (we'll accumulate fresh for next chunk if needed)
					fullContent = ""
					toolCallsAccum = make(map[int]*ollamaToolCallAccumulator)
				}
			} else {
				// Accumulate content for non-streaming
				fullContent += resp.Message.Content
				fullResponse = &resp
				// Also accumulate tool calls
				if len(resp.Message.ToolCalls) > 0 {
					for _, toolCall := range resp.Message.ToolCalls {
						idx := 0
						if toolCall.Function.Index > 0 {
							idx = int(toolCall.Function.Index)
						}

						accum, exists := toolCallsAccum[idx]
						if !exists {
							accum = &ollamaToolCallAccumulator{}
							toolCallsAccum[idx] = accum
						}

						if toolCall.ID != "" {
							accum.id = toolCall.ID
						}
						if toolCall.Function.Name != "" {
							accum.name = toolCall.Function.Name
						}
						if toolCall.Function.Arguments != nil {
							accum.arguments = toolCall.Function.Arguments
						}
					}
				}
			}
			return nil
		})

		if err != nil {
			yield(nil, fmt.Errorf("ollama chat request failed: %w", err))
			return
		}

		// For non-streaming mode, yield the accumulated response
		if !stream && fullResponse != nil {
			fullResponse.Message.Content = fullContent
			modelResp := convertOllamaChatResponseToGenAIWithAccumulatedTools(*fullResponse, fullContent, toolCallsAccum)
			modelResp.TurnComplete = true
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

		// Collect all parts
		var textParts []string
		var toolCalls []api.ToolCall
		var toolResponses []*genai.FunctionResponse

		for _, part := range content.Parts {
			if part == nil {
				continue
			}

			// Handle text content
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}

			// Handle function calls - convert to Ollama tool call format
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, api.ToolCall{
					ID: part.FunctionCall.ID,
					Function: api.ToolCallFunction{
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
					},
				})
			}

			// Handle function responses
			if part.FunctionResponse != nil {
				toolResponses = append(toolResponses, part.FunctionResponse)
			}
		}

		// Join text parts
		contentText := strings.Join(textParts, "\n")

		// Add messages based on what we have
		// For assistant messages with tool calls, include them
		if len(toolCalls) > 0 && (role == "assistant" || role == "model") {
			// Create assistant message with tool calls
			messages = append(messages, api.Message{
				Role:      role,
				Content:   contentText,
				ToolCalls: toolCalls,
			})
		} else if len(toolResponses) > 0 {
			// Tool response messages - add a message for each response
			// These typically come with role="user" or "tool"
			for _, toolResp := range toolResponses {
				// The tool result should be added as a user message with the result
				respJSON := ""
				if toolResp.Response != nil {
					// Convert response to JSON
					respBytes, err := json.Marshal(toolResp.Response)
					if err == nil {
						respJSON = string(respBytes)
					}
				}
				messages = append(messages, api.Message{
					Role:    "user",
					Content: respJSON,
				})
			}
		} else if contentText != "" {
			// Regular text message
			messages = append(messages, api.Message{
				Role:    role,
				Content: contentText,
			})
		}
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
				Required:   []string{},
			}

			// Convert parameters if available
			if funcDecl.Parameters != nil && funcDecl.Parameters.Properties != nil {
				for propName, prop := range funcDecl.Parameters.Properties {
					if prop != nil {
						params.Properties[propName] = convertSchemaPropertyToTool(prop)
					}
				}
				// Only set Required if it's not empty
				if len(funcDecl.Parameters.Required) > 0 {
					params.Required = funcDecl.Parameters.Required
				}
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
			// Ensure we always have a non-nil args map
			args := toolCall.Function.Arguments
			if args == nil {
				args = make(map[string]any)
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
		Partial:      isStreaming && !resp.Done,
		TurnComplete: resp.Done,
		FinishReason: stringToFinishReason(resp.DoneReason),
	}

	return modelResp
}

// convertOllamaChatResponseToGenAIWithAccumulatedTools converts Ollama response with accumulated tool calls
// This is used in streaming mode where tool calls may span multiple chunks
func convertOllamaChatResponseToGenAIWithAccumulatedTools(
	resp api.ChatResponse,
	content string,
	toolCallsAccum map[int]*ollamaToolCallAccumulator,
) *model.LLMResponse {
	// Create the response content with accumulated text
	parts := []*genai.Part{
		{
			Text: content,
		},
	}

	// Add accumulated tool calls
	if len(toolCallsAccum) > 0 {
		// Sort by index to maintain order
		for idx := 0; idx < len(toolCallsAccum); idx++ {
			if accum, exists := toolCallsAccum[idx]; exists {
				args := accum.arguments
				if args == nil {
					args = make(map[string]any)
				}

				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						ID:   accum.id,
						Name: accum.name,
						Args: args,
					},
				})
			}
		}
	}

	// Create the response content
	responseContent := &genai.Content{
		Role:  "model",
		Parts: parts,
	}

	// Create the LLMResponse - mark as complete when done
	modelResp := &model.LLMResponse{
		Content:      responseContent,
		Partial:      false,
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

// mapOllamaModelName maps internal model IDs to actual Ollama model names
// Internal IDs use hyphens, but Ollama model names may use colons (e.g., gpt-oss:20b)
func mapOllamaModelName(modelID string) string {
	// Map of internal IDs to actual Ollama model names
	modelMap := map[string]string{
		"gpt-oss-20b": "gpt-oss:20b",
	}

	// Return mapped name if it exists, otherwise return the original
	if mappedName, exists := modelMap[modelID]; exists {
		return mappedName
	}
	return modelID
}
