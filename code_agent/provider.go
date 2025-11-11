// Package main - Provider definitions and utilities for model backends
package main

import (
	"sort"
	"strings"
)

// Provider represents a backend provider for LLMs
type Provider string

const (
	ProviderGemini   Provider = "gemini"
	ProviderVertexAI Provider = "vertexai"
)

// ProviderMetadata describes a provider and its configuration
type ProviderMetadata struct {
	Name         string   // e.g., "gemini"
	DisplayName  string   // e.g., "Gemini API"
	Icon         string   // e.g., "🔷"
	Description  string   // e.g., "REST API with Google's Gemini models"
	Requirements []string // e.g., ["GOOGLE_API_KEY"]
	IsConfigured bool     // Whether the provider has required environment variables set
}

// String returns the provider name as a string
func (p Provider) String() string {
	return string(p)
}

// AllProviders returns a list of all supported providers
func AllProviders() []Provider {
	return []Provider{
		ProviderGemini,
		ProviderVertexAI,
	}
}

// SortedProviders returns providers in a consistent alphabetical order
func SortedProviders() []Provider {
	providers := AllProviders()
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].String() < providers[j].String()
	})
	return providers
}

// GetProviderMetadata returns metadata about a provider
func GetProviderMetadata(provider Provider) ProviderMetadata {
	switch provider {
	case ProviderGemini:
		return ProviderMetadata{
			Name:        "gemini",
			DisplayName: "Gemini API",
			Icon:        "🔷",
			Description: "REST API with Google's Gemini models",
			Requirements: []string{
				"GOOGLE_API_KEY",
			},
		}
	case ProviderVertexAI:
		return ProviderMetadata{
			Name:        "vertexai",
			DisplayName: "Vertex AI",
			Icon:        "🔶",
			Description: "GCP-native endpoint for Google's Gemini models",
			Requirements: []string{
				"GOOGLE_CLOUD_PROJECT",
				"GOOGLE_CLOUD_LOCATION",
			},
		}
	default:
		return ProviderMetadata{}
	}
}

// ParseProvider converts a string to a Provider, returns empty string if not found
func ParseProvider(s string) Provider {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, p := range AllProviders() {
		if p.String() == s {
			return p
		}
	}
	return Provider("")
}

// IsValidProvider checks if a provider name is valid
func IsValidProvider(name string) bool {
	return ParseProvider(name) != ""
}
