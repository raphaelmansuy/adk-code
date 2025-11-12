package models

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

// OpenAIModelAdapter implements the model.LLM interface for OpenAI models
type OpenAIModelAdapter struct {
	client    openai.Client
	modelName string
}

// createOpenAIModelInternal creates a model using the OpenAI API backend (internal implementation)
func createOpenAIModelInternal(ctx context.Context, cfg OpenAIConfig) (model.LLM, error) {
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
