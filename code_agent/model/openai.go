// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type OpenAIConfig struct {
	APIKey    string
	ModelName string
}

// OpenAIModelAdapter implements the model.LLM interface for OpenAI models
type OpenAIModelAdapter struct {
	client    openai.Client
	modelName string
}

// CreateOpenAIModel creates a model using the OpenAI API backend
func CreateOpenAIModel(ctx context.Context, cfg OpenAIConfig) (model.LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if cfg.ModelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	client := openai.NewClient(option.WithAPIKey(cfg.APIKey))

	return &OpenAIModelAdapter{
		client:    client,
		modelName: cfg.ModelName,
	}, nil
}

// Name returns the model name
func (a *OpenAIModelAdapter) Name() string {
	return a.modelName
}

// isReasoningModel checks if the model is an o-series reasoning model
// These models have different parameter requirements (no temperature, top_p, etc.)
func isReasoningModel(modelName string) bool {
	// O-series models: o1, o1-preview, o1-mini, o3, o3-mini, etc.
	return strings.HasPrefix(modelName, "o1") ||
		strings.HasPrefix(modelName, "o3") ||
		strings.Contains(modelName, "gpt-5") // gpt-5-nano is likely an o-series model
}

// GenerateContent implements the model.LLM interface
// It handles both streaming and non-streaming requests
func (a *OpenAIModelAdapter) GenerateContent(
	ctx context.Context,
	req *model.LLMRequest,
	stream bool,
) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// Convert genai.Content to OpenAI chat completion messages
		messages, err := convertToOpenAIMessages(req.Contents)
		if err != nil {
			yield(nil, fmt.Errorf("failed to convert contents to OpenAI messages: %w", err))
			return
		}

		// Extract model name from request if not set
		modelName := a.modelName
		if req.Model != "" {
			modelName = req.Model
		}

		// Prepare OpenAI request
		openaiReq := openai.ChatCompletionNewParams{
			Model:    modelName,
			Messages: messages,
		}

		// Apply config if provided
		// Note: O-series reasoning models (o1, o3, etc.) don't support temperature, top_p, or stop sequences
		isReasoning := isReasoningModel(modelName)

		if req.Config != nil {
			// Temperature and TopP are not supported by reasoning models
			if !isReasoning {
				if req.Config.Temperature != nil {
					openaiReq.Temperature = param.NewOpt(float64(*req.Config.Temperature))
				}
				if req.Config.TopP != nil {
					openaiReq.TopP = param.NewOpt(float64(*req.Config.TopP))
				}
			}

			// Max tokens is supported by all models
			if req.Config.MaxOutputTokens > 0 {
				openaiReq.MaxCompletionTokens = param.NewOpt(int64(req.Config.MaxOutputTokens))
			}

			// Stop sequences not supported by reasoning models
			if !isReasoning && len(req.Config.StopSequences) > 0 {
				openaiReq.Stop = openai.ChatCompletionNewParamsStopUnion{
					OfStringArray: req.Config.StopSequences,
				}
			}
		}

		// Configure tool calling if tools are provided in Config
		if req.Config != nil && len(req.Config.Tools) > 0 {
			tools, err := convertToOpenAITools(req.Config.Tools)
			if err != nil {
				yield(nil, fmt.Errorf("failed to convert tools: %w", err))
				return
			}
			// Only set tools and tool_choice if we actually have valid converted tools
			if len(tools) > 0 {
				openaiReq.Tools = tools

				// Map tool choice from genai.ToolConfig to OpenAI format
				// genai modes: ModeAuto (model decides), ModeAny (required), ModeNone (disable)
				// OpenAI modes: "auto", "required" (via allowed_tools), "none" (via string)
				if req.Config.ToolConfig != nil && req.Config.ToolConfig.FunctionCallingConfig != nil {
					switch req.Config.ToolConfig.FunctionCallingConfig.Mode {
					case genai.FunctionCallingConfigModeAny:
						// Force model to call at least one tool
						// Use the "required" mode via allowed_tools
						openaiReq.ToolChoice = openai.ToolChoiceOptionAllowedTools(openai.ChatCompletionAllowedToolsParam{
							Mode:  openai.ChatCompletionAllowedToolsModeRequired,
							Tools: convertToolsToMaps(tools),
						})
					case genai.FunctionCallingConfigModeNone:
						// Disable tool calling - use "none" string
						openaiReq.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
							OfAuto: param.NewOpt("none"),
						}
					case genai.FunctionCallingConfigModeAuto, genai.FunctionCallingConfigModeUnspecified:
						// Model decides (default) - use "auto" string
						openaiReq.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
							OfAuto: param.NewOpt("auto"),
						}
					default:
						// Default to auto for unknown modes
						openaiReq.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
							OfAuto: param.NewOpt("auto"),
						}
					}
				} else {
					// No tool config specified, default to "auto"
					openaiReq.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
						OfAuto: param.NewOpt("auto"),
					}
				}
			}
		}

		// Handle streaming vs non-streaming
		if stream {
			// Streaming path
			streamResp := a.client.Chat.Completions.NewStreaming(ctx, openaiReq)

			// Track accumulated tool calls across streaming deltas
			// OpenAI streams tool calls incrementally with an index
			type toolCallAccumulator struct {
				id        string
				name      string
				arguments string // accumulated JSON fragments
			}
			toolCallsAccum := make(map[int]*toolCallAccumulator)

			// Process stream events
			for streamResp.Next() {
				event := streamResp.Current()
				if len(event.Choices) > 0 {
					choice := event.Choices[0]

					// Convert to LLMResponse
					resp := &model.LLMResponse{
						Partial: true,
					}

					// Build content with both text and tool calls
					content := &genai.Content{
						Role:  "model",
						Parts: []*genai.Part{},
					}

					// Extract text content from delta
					if choice.Delta.Content != "" {
						content.Parts = append(content.Parts, &genai.Part{
							Text: choice.Delta.Content,
						})
					}

					// Accumulate tool call deltas
					// OpenAI sends tool calls with an Index field to track which call is being updated
					for _, toolCall := range choice.Delta.ToolCalls {
						if toolCall.Type == "function" {
							idx := toolCall.Index

							// Get or create accumulator for this index
							accum, exists := toolCallsAccum[int(idx)]
							if !exists {
								accum = &toolCallAccumulator{}
								toolCallsAccum[int(idx)] = accum
							}

							// Accumulate fields (they may arrive in separate deltas)
							if toolCall.ID != "" {
								accum.id = toolCall.ID
							}
							if toolCall.Function.Name != "" {
								accum.name = toolCall.Function.Name
							}
							if toolCall.Function.Arguments != "" {
								accum.arguments += toolCall.Function.Arguments
							}
						}
					}

					// On finish reason, parse accumulated tool calls
					if choice.FinishReason != "" {
						// Parse all accumulated tool calls now that we have complete JSON
						for _, accum := range toolCallsAccum {
							var args map[string]any
							if accum.arguments != "" {
								var argsData interface{}
								if err := json.Unmarshal([]byte(accum.arguments), &argsData); err == nil {
									if argsMap, ok := argsData.(map[string]interface{}); ok {
										args = argsMap
									}
								}
								// If parsing fails, args remains nil - tool will receive empty args
							}

							content.Parts = append(content.Parts, &genai.Part{
								FunctionCall: &genai.FunctionCall{
									Name: accum.name,
									Args: args,
									ID:   accum.id,
								},
							})
						}
					}

					// Only set content if we have parts
					if len(content.Parts) > 0 {
						resp.Content = content
					}

					// Map finish reason if present - note FinishReason is a string
					// When finish reason is set, the turn is complete
					if choice.FinishReason != "" {
						resp.FinishReason = mapFinishReason(choice.FinishReason)
						resp.TurnComplete = true
					}

					if !yield(resp, nil) {
						return
					}
				}
			}

			if err := streamResp.Err(); err != nil {
				yield(nil, fmt.Errorf("stream error: %w", err))
			}
		} else {
			// Non-streaming path
			completion, err := a.client.Chat.Completions.New(ctx, openaiReq)
			if err != nil {
				yield(nil, fmt.Errorf("failed to create completion: %w", err))
				return
			}

			// Convert OpenAI response to genai response
			resp, err := convertFromOpenAICompletion(completion)
			if err != nil {
				yield(nil, fmt.Errorf("failed to convert OpenAI response: %w", err))
				return
			}

			resp.TurnComplete = true
			if !yield(resp, nil) {
				return
			}
		}
	}
}

// convertToOpenAIMessages converts genai.Content slice to OpenAI chat completion messages
// Supports text, function calls, and function responses
func convertToOpenAIMessages(contents []*genai.Content) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(contents))

	for _, content := range contents {
		if content == nil {
			continue
		}

		roleStr := strings.ToLower(content.Role)

		// Collect text content and check for function calls/responses
		var textContent string
		var functionCalls []openai.ChatCompletionMessageToolCallUnionParam
		var functionResponses []genai.FunctionResponse

		for _, part := range content.Parts {
			if part == nil {
				continue
			}

			// Handle text content
			if part.Text != "" {
				if textContent != "" {
					textContent += "\n"
				}
				textContent += part.Text
			}

			// Handle function calls
			if part.FunctionCall != nil {
				// Convert args to JSON string
				argsJSON := ""
				if part.FunctionCall.Args != nil {
					argsBytes, err := json.Marshal(part.FunctionCall.Args)
					if err == nil {
						argsJSON = string(argsBytes)
					}
				}

				functionCalls = append(functionCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: part.FunctionCall.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      part.FunctionCall.Name,
							Arguments: argsJSON,
						},
					},
				})
			}

			// Handle function responses
			if part.FunctionResponse != nil {
				functionResponses = append(functionResponses, *part.FunctionResponse)
			}
		}

		// Create message based on role and content type
		// Handle function responses FIRST (they can appear in any role)
		if len(functionResponses) > 0 {
			// Tool response messages
			for _, funcResp := range functionResponses {
				// Convert response to JSON string
				respJSON := ""
				if funcResp.Response != nil {
					// Convert error types to strings before marshaling
					// The Response map may contain error types which don't marshal properly
					cleanedResponse := make(map[string]any)
					for k, v := range funcResp.Response {
						// Check if the value is an error type and convert to string
						if err, isError := v.(error); isError {
							cleanedResponse[k] = err.Error()
						} else {
							cleanedResponse[k] = v
						}
					}

					// Marshal the cleaned response
					respBytes, err := json.Marshal(cleanedResponse)
					if err == nil {
						respJSON = string(respBytes)
					} else {
						// Fallback to error message if marshaling fails
						respJSON = fmt.Sprintf("{\"error\": \"failed to marshal response: %v\"}", err)
					}
				} else {
					// Empty response
					respJSON = "{}"
				}

				messages = append(messages, openai.ToolMessage(respJSON, funcResp.ID))
			}
			continue // Skip normal role handling
		}

		switch roleStr {
		case "user":
			// User message - can contain text or nothing
			if textContent != "" {
				messages = append(messages, openai.UserMessage(textContent))
			}

		case "assistant", "model":
			// Assistant message - can contain text and/or tool calls
			if len(functionCalls) > 0 {
				// Assistant message with tool calls
				msg := openai.ChatCompletionAssistantMessageParam{
					ToolCalls: functionCalls,
				}
				if textContent != "" {
					msg.Content.OfString = param.NewOpt(textContent)
				}
				messages = append(messages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &msg,
				})
			} else if textContent != "" {
				// Simple text assistant message
				messages = append(messages, openai.AssistantMessage(textContent))
			}

		case "tool", "function":
			// This case is now handled above, but keep for backwards compatibility
			// (should not reach here due to the continue statement above)

		case "system":
			// System message
			if textContent != "" {
				messages = append(messages, openai.SystemMessage(textContent))
			}

		default:
			// Default to user message
			if textContent != "" {
				messages = append(messages, openai.UserMessage(textContent))
			}
		}
	}

	return messages, nil
}

// convertFromOpenAICompletion converts an OpenAI ChatCompletion to genai.LLMResponse
func convertFromOpenAICompletion(completion *openai.ChatCompletion) (*model.LLMResponse, error) {
	resp := &model.LLMResponse{
		Partial: false,
	}

	// Extract first candidate
	if len(completion.Choices) > 0 {
		choice := completion.Choices[0]

		// Map finish reason (FinishReason is a string, not a pointer)
		if choice.FinishReason != "" {
			resp.FinishReason = mapFinishReason(choice.FinishReason)
		}

		// Build content with both text and tool calls
		content := &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{},
		}

		// Extract text content (Content is a string, not a pointer)
		if choice.Message.Content != "" {
			content.Parts = append(content.Parts, &genai.Part{
				Text: choice.Message.Content,
			})
		}

		// Extract tool calls and convert to function calls
		for _, toolCall := range choice.Message.ToolCalls {
			// Check if it's a function tool call (Type should be "function")
			if toolCall.Type == "function" {
				funcCall := toolCall.AsFunction()

				// Parse arguments JSON
				var args map[string]any
				if funcCall.Function.Arguments != "" {
					// Try to parse arguments as JSON
					var argsData interface{}
					if err := json.Unmarshal([]byte(funcCall.Function.Arguments), &argsData); err == nil {
						if argsMap, ok := argsData.(map[string]interface{}); ok {
							args = argsMap
						}
					}
				}

				// Add function call part
				content.Parts = append(content.Parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: funcCall.Function.Name,
						Args: args,
						ID:   funcCall.ID,
					},
				})
			}
		}

		// Only set content if we have parts
		if len(content.Parts) > 0 {
			resp.Content = content
		}
	}

	// Map usage metadata (Usage is a struct, not a pointer)
	resp.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     int32(completion.Usage.PromptTokens),
		CandidatesTokenCount: int32(completion.Usage.CompletionTokens),
		TotalTokenCount:      int32(completion.Usage.TotalTokens),
	}

	return resp, nil
}

// mapFinishReason converts OpenAI finish reason string to genai FinishReason
func mapFinishReason(reason string) genai.FinishReason {
	switch reason {
	case "stop":
		return genai.FinishReasonStop
	case "length":
		return genai.FinishReasonMaxTokens
	case "tool_calls":
		// Map tool_calls to stop (there's no dedicated FinishReasonToolCalls)
		return genai.FinishReasonStop
	case "content_filter":
		return genai.FinishReasonSafety
	case "function_call":
		// Map function_call to stop
		return genai.FinishReasonStop
	default:
		return genai.FinishReasonOther
	}
}

// convertToOpenAITools converts ADK tools to OpenAI tools format
// The tools parameter is expected to be []*genai.Tool containing function declarations
func convertToOpenAITools(tools []*genai.Tool) ([]openai.ChatCompletionToolUnionParam, error) {
	if tools == nil {
		return []openai.ChatCompletionToolUnionParam{}, nil
	}

	var openaiTools []openai.ChatCompletionToolUnionParam

	// Iterate through the genai.Tool slice
	for _, genaiTool := range tools {
		if genaiTool == nil {
			continue
		}

		// Convert function declarations
		if genaiTool.FunctionDeclarations != nil {
			for _, funcDecl := range genaiTool.FunctionDeclarations {
				if funcDecl == nil {
					continue
				}

				// Build OpenAI function definition
				functionDef := openai.FunctionDefinitionParam{
					Name:        funcDecl.Name,
					Description: param.NewOpt(funcDecl.Description),
				}

				// Convert parameters schema
				// ADK uses genai.Schema or JSON Schema format
				if funcDecl.Parameters != nil {
					// Convert genai.Schema to map for OpenAI
					params, err := convertSchemaToMapWithError(funcDecl.Parameters)
					if err != nil {
						return nil, fmt.Errorf("failed to convert schema for function %s: %w", funcDecl.Name, err)
					}
					if params != nil {
						functionDef.Parameters = params
					}
				} else if funcDecl.ParametersJsonSchema != nil {
					// Use JSON Schema directly - convert via JSON marshaling
					// Try direct map type assertions first
					if params, ok := funcDecl.ParametersJsonSchema.(map[string]interface{}); ok {
						functionDef.Parameters = params
					} else if params, ok := funcDecl.ParametersJsonSchema.(map[string]any); ok {
						functionDef.Parameters = params
					} else {
						// Type is not a map - try to convert via JSON marshaling
						// This handles types like *jsonschema.Schema
						schemaBytes, err := json.Marshal(funcDecl.ParametersJsonSchema)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal ParametersJsonSchema for function %s: %w", funcDecl.Name, err)
						}

						var params map[string]interface{}
						if err := json.Unmarshal(schemaBytes, &params); err != nil {
							return nil, fmt.Errorf("failed to unmarshal ParametersJsonSchema for function %s: %w", funcDecl.Name, err)
						}
						functionDef.Parameters = params
					}
				}

				// Create OpenAI tool using the helper function
				openaiTool := openai.ChatCompletionFunctionTool(functionDef)

				openaiTools = append(openaiTools, openaiTool)
			}
		}
	}

	return openaiTools, nil
}

// convertSchemaToMapWithError converts a genai.Schema to a map suitable for OpenAI
// Returns error if conversion fails (e.g., circular references)
func convertSchemaToMapWithError(schema *genai.Schema) (map[string]interface{}, error) {
	if schema == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	// Add type
	if schema.Type != "" {
		result["type"] = strings.ToLower(string(schema.Type))
	}

	// Add description
	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Add properties for object types
	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, propSchema := range schema.Properties {
			converted, err := convertSchemaToMapWithError(propSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to convert property %s: %w", name, err)
			}
			props[name] = converted
		}
		result["properties"] = props
	}

	// Add required fields
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// Add items for array types
	if schema.Items != nil {
		converted, err := convertSchemaToMapWithError(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array items: %w", err)
		}
		result["items"] = converted
	}

	// Add enum values
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	// Add format
	if schema.Format != "" {
		result["format"] = schema.Format
	}

	// Add numeric constraints
	if schema.Minimum != nil {
		result["minimum"] = *schema.Minimum
	}
	if schema.Maximum != nil {
		result["maximum"] = *schema.Maximum
	}

	// Add string constraints
	if schema.MinLength != nil {
		result["minLength"] = *schema.MinLength
	}
	if schema.MaxLength != nil {
		result["maxLength"] = *schema.MaxLength
	}

	// Add pattern
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}

	return result, nil
}

// convertToolsToMaps converts OpenAI tool union params to map format for allowed_tools
// This is needed for the "required" tool choice mode
func convertToolsToMaps(tools []openai.ChatCompletionToolUnionParam) []map[string]any {
	result := make([]map[string]any, len(tools))
	for i, tool := range tools {
		// Convert each tool to a map by marshaling and unmarshaling
		// This preserves the tool structure in the format OpenAI expects
		toolBytes, err := json.Marshal(tool)
		if err != nil {
			// If marshal fails, create a minimal map
			result[i] = map[string]any{"type": "function"}
			continue
		}

		var toolMap map[string]any
		if err := json.Unmarshal(toolBytes, &toolMap); err != nil {
			// If unmarshal fails, create a minimal map
			result[i] = map[string]any{"type": "function"}
			continue
		}

		result[i] = toolMap
	}
	return result
}
