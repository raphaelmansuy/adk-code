package tools

// ToolExecutionListener receives notifications about tool execution events.
// This interface decouples tool execution from display rendering, allowing
// different implementations to handle tool lifecycle events.
type ToolExecutionListener interface {
	// OnToolStart is called when a tool execution begins
	OnToolStart(toolName string, input interface{})

	// OnToolProgress is called during tool execution with progress updates
	OnToolProgress(toolName string, stage string, progress string)

	// OnToolComplete is called when tool execution finishes
	OnToolComplete(toolName string, result interface{}, err error)
}

// DefaultToolExecutionListener is a no-op implementation of ToolExecutionListener
type DefaultToolExecutionListener struct{}

// OnToolStart is a no-op implementation
func (l *DefaultToolExecutionListener) OnToolStart(toolName string, input interface{}) {
	// No-op
}

// OnToolProgress is a no-op implementation
func (l *DefaultToolExecutionListener) OnToolProgress(toolName string, stage string, progress string) {
	// No-op
}

// OnToolComplete is a no-op implementation
func (l *DefaultToolExecutionListener) OnToolComplete(toolName string, result interface{}, err error) {
	// No-op
}

// NewDefaultToolExecutionListener creates a no-op tool execution listener
func NewDefaultToolExecutionListener() ToolExecutionListener {
	return &DefaultToolExecutionListener{}
}

// ToolRendererAdapter wraps ToolRenderer to implement ToolExecutionListener
// This adapter allows ToolRenderer to be used as a listener without modifying
// the original ToolRenderer implementation.
type ToolRendererAdapter struct {
	renderer *ToolRenderer
}

// NewToolRendererAdapter creates a new adapter for ToolRenderer
func NewToolRendererAdapter(renderer *ToolRenderer) *ToolRendererAdapter {
	return &ToolRendererAdapter{
		renderer: renderer,
	}
}

// OnToolStart implements ToolExecutionListener
func (a *ToolRendererAdapter) OnToolStart(toolName string, input interface{}) {
	if inputMap, ok := input.(map[string]any); ok {
		a.renderer.RenderToolExecution(toolName, inputMap)
	}
}

// OnToolProgress implements ToolExecutionListener
func (a *ToolRendererAdapter) OnToolProgress(toolName string, stage string, progress string) {
	// Progress events could render a status update
	// This is a placeholder for future implementation
}

// OnToolComplete implements ToolExecutionListener
func (a *ToolRendererAdapter) OnToolComplete(toolName string, result interface{}, err error) {
	if resultMap, ok := result.(map[string]any); ok {
		a.renderer.RenderToolResultDetailed(toolName, resultMap)
	}
}
