package display

import (
	"fmt"
	"strings"
	"sync"
)

// MessageType represents different types of agent messages
type MessageType int

const (
	MessageTypeThinking MessageType = iota
	MessageTypeResponse
	MessageTypeTool
	MessageTypeToolResult
	MessageTypeCompletion
	MessageTypeError
)

// StreamingSegment handles streaming display of a message segment
type StreamingSegment struct {
	mu             sync.Mutex
	messageType    MessageType
	buffer         strings.Builder
	frozen         bool
	mdRenderer     *MarkdownRenderer
	typewriter     *TypewriterPrinter
	outputFormat   string
	headerRendered bool
}

// NewStreamingSegment creates a new streaming segment
func NewStreamingSegment(msgType MessageType, mdRenderer *MarkdownRenderer, typewriter *TypewriterPrinter, outputFormat string) *StreamingSegment {
	ss := &StreamingSegment{
		messageType:  msgType,
		mdRenderer:   mdRenderer,
		typewriter:   typewriter,
		outputFormat: outputFormat,
	}

	// Render header immediately when creating segment
	if IsTTY() && outputFormat != OutputFormatPlain {
		ss.renderHeader()
	}

	return ss
}

// AppendText adds text to the buffer (does not render immediately)
func (ss *StreamingSegment) AppendText(text string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.frozen {
		return
	}

	ss.buffer.WriteString(text)
}

// Freeze finalizes the segment and renders the accumulated content
func (ss *StreamingSegment) Freeze() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.frozen {
		return
	}

	ss.frozen = true
	content := ss.buffer.String()

	if content != "" {
		ss.renderContent(content)
	}
}

// renderHeader renders the contextual header for this segment type
func (ss *StreamingSegment) renderHeader() {
	var header string

	switch ss.messageType {
	case MessageTypeThinking:
		header = "### Agent is thinking\n"
	case MessageTypeResponse:
		header = "### Agent responds\n"
	case MessageTypeTool:
		header = "### Tool execution\n"
	case MessageTypeToolResult:
		header = "### Tool result\n"
	case MessageTypeCompletion:
		header = "### Task complete\n"
	case MessageTypeError:
		header = "### Error\n"
	default:
		header = "### Message\n"
	}

	// Render markdown header
	if ss.mdRenderer != nil {
		rendered, err := ss.mdRenderer.Render(header)
		if err == nil {
			fmt.Print("\n" + rendered)
		} else {
			fmt.Print("\n" + header)
		}
	} else {
		fmt.Print("\n" + header)
	}

	ss.headerRendered = true
}

// renderContent renders the accumulated content
func (ss *StreamingSegment) renderContent(content string) {
	var output string

	// Render based on message type
	switch ss.messageType {
	case MessageTypeThinking, MessageTypeResponse:
		// Render as markdown
		if ss.mdRenderer != nil && ss.outputFormat != OutputFormatPlain {
			rendered, err := ss.mdRenderer.Render(content)
			if err == nil {
				output = rendered
			} else {
				output = content
			}
		} else {
			output = content
		}

	case MessageTypeTool, MessageTypeToolResult:
		// Tool messages are already formatted
		output = content

	case MessageTypeCompletion:
		// Render as markdown
		if ss.mdRenderer != nil && ss.outputFormat != OutputFormatPlain {
			rendered, err := ss.mdRenderer.Render(content)
			if err == nil {
				output = rendered
			} else {
				output = content
			}
		} else {
			output = content
		}

	case MessageTypeError:
		// Error messages rendered as-is
		output = content

	default:
		output = content
	}

	// Print with typewriter if enabled, otherwise print normally
	if output != "" {
		if ss.typewriter != nil && ss.typewriter.IsEnabled() {
			ss.typewriter.Print(output)
		} else {
			fmt.Print(output)
		}

		// Ensure newline at end
		if !strings.HasSuffix(output, "\n") {
			fmt.Println()
		}
	}
}

// IsFrozen returns whether the segment is frozen
func (ss *StreamingSegment) IsFrozen() bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.frozen
}
