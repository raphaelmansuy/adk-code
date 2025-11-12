package tools

import (
	"testing"

	"code_agent/internal/display/renderer"
	"code_agent/internal/display/styles"
)

// Re-export for tests
var (
	NewRenderer       = renderer.NewRenderer
	OutputFormatPlain = styles.OutputFormatPlain
)

func TestDefaultToolExecutionListener(t *testing.T) {
	listener := NewDefaultToolExecutionListener()

	// These should not panic
	listener.OnToolStart("test_tool", map[string]any{"param": "value"})
	listener.OnToolProgress("test_tool", "stage1", "progress1")
	listener.OnToolComplete("test_tool", map[string]any{"result": "success"}, nil)
}

func TestToolExecutionListenerInterface(t *testing.T) {
	var _ ToolExecutionListener = NewDefaultToolExecutionListener()
	adapter := NewToolRendererAdapter(nil)
	var _ ToolExecutionListener = adapter
}

func TestToolRendererAdapter(t *testing.T) {
	renderer, _ := NewRenderer(OutputFormatPlain)
	toolRenderer := NewToolRenderer(renderer)
	adapter := NewToolRendererAdapter(toolRenderer)

	if adapter == nil {
		t.Error("NewToolRendererAdapter should not return nil")
	}

	// Test that adapter doesn't panic with nil result
	adapter.OnToolComplete("test", nil, nil)
	adapter.OnToolProgress("test", "stage", "progress")
}

func TestMultipleListeners(t *testing.T) {
	listener1 := NewDefaultToolExecutionListener()
	listener2 := NewDefaultToolExecutionListener()

	// Both should work independently
	listener1.OnToolStart("tool1", map[string]any{})
	listener2.OnToolStart("tool2", map[string]any{})

	listener1.OnToolProgress("tool1", "s1", "p1")
	listener2.OnToolProgress("tool2", "s2", "p2")

	listener1.OnToolComplete("tool1", map[string]any{}, nil)
	listener2.OnToolComplete("tool2", map[string]any{}, nil)
}

func TestToolRendererAdapterWithMapInput(t *testing.T) {
	renderer, _ := NewRenderer(OutputFormatPlain)
	toolRenderer := NewToolRenderer(renderer)
	adapter := NewToolRendererAdapter(toolRenderer)

	input := map[string]any{"file": "test.txt", "lines": 10}
	adapter.OnToolStart("read_file", input)

	result := map[string]any{"content": "test content", "bytes": 12}
	adapter.OnToolComplete("read_file", result, nil)
}
