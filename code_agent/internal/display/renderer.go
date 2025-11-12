package display

import (
	"code_agent/internal/display/components"
	"code_agent/internal/display/formatters"
	"code_agent/internal/display/styles"
)

// Re-export types and constants for backward compatibility
type (
	EventType     = components.EventType
	TimelineEvent = components.TimelineEvent
	EventTimeline = components.EventTimeline
	APIUsageInfo  = formatters.APIUsageInfo
)

// Re-export functions
var (
	EventTypeIcon    = components.EventTypeIcon
	NewEventTimeline = components.NewEventTimeline
)

// OutputFormat constants re-export
const (
	OutputFormatRich  = styles.OutputFormatRich
	OutputFormatPlain = styles.OutputFormatPlain
	OutputFormatJSON  = styles.OutputFormatJSON
)

// Event types - re-export for backward compatibility
const (
	EventTypeThinking  = components.EventTypeThinking
	EventTypeExecuting = components.EventTypeExecuting
	EventTypeResult    = components.EventTypeResult
	EventTypeSuccess   = components.EventTypeSuccess
	EventTypeWarning   = components.EventTypeWarning
	EventTypeError     = components.EventTypeError
	EventTypeProgress  = components.EventTypeProgress
)

// Note: Renderer alias and NewRenderer are provided by the top-level display facade.
