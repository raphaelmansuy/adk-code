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

package formatters

import (
	"fmt"
	"sync"

	"code_agent/display/styles"
)

// Formatter is a common interface for all formatter types
// This allows for polymorphic handling of different formatters
type Formatter interface {
	// Type returns the formatter type identifier
	Type() string
}

// FormatterRegistry manages all available formatters
// This enables centralized formatter lifecycle and extensibility
type FormatterRegistry struct {
	mu               sync.RWMutex
	agentFormatter   *AgentFormatter
	toolFormatter    *ToolFormatter
	errorFormatter   *ErrorFormatter
	metricsFormatter *MetricsFormatter
	customFormatters map[string]Formatter
}

// NewFormatterRegistry creates a new formatter registry with all default formatters
func NewFormatterRegistry(outputFormat string, s *styles.Styles, f *styles.Formatter, mdRenderer MarkdownRenderer) *FormatterRegistry {
	return &FormatterRegistry{
		agentFormatter:   NewAgentFormatter(outputFormat, s, f, mdRenderer),
		toolFormatter:    NewToolFormatter(outputFormat, s, f),
		errorFormatter:   NewErrorFormatter(outputFormat, s, f, mdRenderer),
		metricsFormatter: NewMetricsFormatter(outputFormat, s, f, mdRenderer),
		customFormatters: make(map[string]Formatter),
	}
}

// GetAgentFormatter returns the agent formatter
func (fr *FormatterRegistry) GetAgentFormatter() *AgentFormatter {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return fr.agentFormatter
}

// GetToolFormatter returns the tool formatter
func (fr *FormatterRegistry) GetToolFormatter() *ToolFormatter {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return fr.toolFormatter
}

// GetErrorFormatter returns the error formatter
func (fr *FormatterRegistry) GetErrorFormatter() *ErrorFormatter {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return fr.errorFormatter
}

// GetMetricsFormatter returns the metrics formatter
func (fr *FormatterRegistry) GetMetricsFormatter() *MetricsFormatter {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return fr.metricsFormatter
}

// RegisterCustomFormatter adds a custom formatter to the registry
// This enables extensibility without modifying core formatters
func (fr *FormatterRegistry) RegisterCustomFormatter(name string, formatter Formatter) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if _, exists := fr.customFormatters[name]; exists {
		return fmt.Errorf("formatter already registered: %s", name)
	}

	fr.customFormatters[name] = formatter
	return nil
}

// GetCustomFormatter retrieves a custom formatter by name
func (fr *FormatterRegistry) GetCustomFormatter(name string) (Formatter, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	formatter, ok := fr.customFormatters[name]
	if !ok {
		return nil, fmt.Errorf("formatter not found: %s", name)
	}

	return formatter, nil
}

// ListCustomFormatters returns all registered custom formatter names
func (fr *FormatterRegistry) ListCustomFormatters() []string {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	names := make([]string, 0, len(fr.customFormatters))
	for name := range fr.customFormatters {
		names = append(names, name)
	}
	return names
}
