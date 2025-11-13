package formatters

import (
	"sync"

	"adk-code/internal/display/styles"
	pkgerrors "adk-code/pkg/errors"
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
		return pkgerrors.InvalidInputError("formatter already registered: " + name)
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
		return nil, pkgerrors.InvalidInputError("formatter not found: " + name)
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
