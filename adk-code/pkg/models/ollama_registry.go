// Package models - Ollama model registry with caching
package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// OllamaModelRegistry manages Ollama model discovery with caching
type OllamaModelRegistry struct {
	client *api.Client
	mu     sync.RWMutex
	cache  *modelCache
}

// modelCache stores cached model information with expiration
type modelCache struct {
	models    []ModelInfo
	timestamp time.Time
	ttl       time.Duration
}

// NewOllamaModelRegistry creates a new Ollama model registry
func NewOllamaModelRegistry(client *api.Client, cacheTTL time.Duration) *OllamaModelRegistry {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute // Default 5 minute TTL
	}

	return &OllamaModelRegistry{
		client: client,
		cache: &modelCache{
			ttl: cacheTTL,
		},
	}
}

// ListModels returns a list of available Ollama models, using cache if valid
func (r *OllamaModelRegistry) ListModels(ctx context.Context, forceRefresh bool) ([]ModelInfo, error) {
	r.mu.RLock()
	// Check if cache is valid and we're not forcing a refresh
	if !forceRefresh && r.cache.models != nil && time.Since(r.cache.timestamp) < r.cache.ttl {
		models := r.cache.models
		r.mu.RUnlock()
		return models, nil
	}
	r.mu.RUnlock()

	// Cache is invalid or refresh requested, fetch fresh data
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have updated)
	if !forceRefresh && r.cache.models != nil && time.Since(r.cache.timestamp) < r.cache.ttl {
		return r.cache.models, nil
	}

	// Fetch models from Ollama API
	listResp, err := r.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Ollama models: %w", err)
	}

	// Convert to ModelInfo
	models := make([]ModelInfo, 0, len(listResp.Models))
	for _, model := range listResp.Models {
		modelInfo := r.convertOllamaModelToInfo(model)
		models = append(models, modelInfo)
	}

	// Update cache
	r.cache.models = models
	r.cache.timestamp = time.Now()

	return models, nil
}

// GetModelInfo retrieves detailed information about a specific model
func (r *OllamaModelRegistry) GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error) {
	// First try to find in cached list
	models, err := r.ListModels(ctx, false)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.Name == modelName {
			// Fetch detailed info using Show API
			return r.fetchDetailedModelInfo(ctx, modelName, &model)
		}
	}

	// Model not found in cache, try fetching directly
	return r.fetchDetailedModelInfo(ctx, modelName, nil)
}

// fetchDetailedModelInfo fetches detailed model information using the Show API
func (r *OllamaModelRegistry) fetchDetailedModelInfo(ctx context.Context, modelName string, baseInfo *ModelInfo) (*ModelInfo, error) {
	showReq := &api.ShowRequest{
		Name: modelName,
	}

	showResp, err := r.client.Show(ctx, showReq)
	if err != nil {
		// If we have base info, return it even if Show fails
		if baseInfo != nil {
			return baseInfo, nil
		}
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}

	// Create or update ModelInfo with detailed information
	info := baseInfo
	if info == nil {
		info = &ModelInfo{
			Name:     modelName,
			Provider: "ollama",
		}
	}

	// Update with detailed information from Show response
	if showResp.ModelInfo != nil {
		// Extract family from model info
		if family, ok := showResp.ModelInfo["general.architecture"].(string); ok {
			info.Family = family
		}

		// Extract parameter count
		if paramCount, ok := showResp.ModelInfo["general.parameter_count"].(float64); ok {
			info.ParameterCount = int64(paramCount)
		}

		// Infer size from parameter count
		if info.ParameterCount > 0 {
			info.Size = formatParameterSize(info.ParameterCount)
		}
	}

	// Update capabilities based on detailed info
	if info.Family != "" {
		info.Capabilities = DetectCapabilitiesFromName(modelName, info.Family)
	}

	return info, nil
}

// convertOllamaModelToInfo converts an Ollama API model to ModelInfo
func (r *OllamaModelRegistry) convertOllamaModelToInfo(model api.ListModelResponse) ModelInfo {
	// Extract model name and details
	name := model.Name
	family := model.Details.Family
	size := formatSize(model.Size)

	// Detect capabilities
	caps := DetectCapabilitiesFromName(name, family)

	info := ModelInfo{
		Name:         name,
		DisplayName:  name,
		Provider:     "ollama",
		Family:       family,
		Size:         size,
		Quantization: model.Details.QuantizationLevel,
		Capabilities: caps,
		Modified:     model.ModifiedAt.Format(time.RFC3339),
	}

	return info
}

// InvalidateCache forces the cache to be refreshed on next ListModels call
func (r *OllamaModelRegistry) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache.timestamp = time.Time{} // Set to zero time to invalidate
}

// formatSize formats a byte size into a human-readable string
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatParameterSize formats parameter count into a readable string (e.g., "7B", "13B")
func formatParameterSize(params int64) string {
	const billion = 1000000000
	if params >= billion {
		return fmt.Sprintf("%.1fB", float64(params)/float64(billion))
	}
	return fmt.Sprintf("%dM", params/1000000)
}
