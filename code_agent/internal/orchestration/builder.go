package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"

	"code_agent/internal/config"
)

// Orchestrator is a builder for Application components
// It provides a fluent API for creating and configuring all application components
type Orchestrator struct {
	ctx               context.Context
	cfg               *config.Config
	displayComponents *DisplayComponents
	modelComponents   *ModelComponents
	agentComponent    agent.Agent
	mcpComponents     *MCPComponents
	sessionComponents *SessionComponents
	err               error
}

// NewOrchestrator creates a new Orchestrator for building application components
func NewOrchestrator(ctx context.Context, cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		ctx: ctx,
		cfg: cfg,
	}
}

// WithDisplay initializes display components
func (o *Orchestrator) WithDisplay() *Orchestrator {
	if o.err != nil {
		return o
	}
	o.displayComponents, o.err = InitializeDisplayComponents(o.cfg)
	return o
}

// WithModel initializes model/LLM components
func (o *Orchestrator) WithModel() *Orchestrator {
	if o.err != nil {
		return o
	}
	o.modelComponents, o.err = InitializeModelComponents(o.ctx, o.cfg)
	return o
}

// WithAgent initializes the agent component
func (o *Orchestrator) WithAgent() *Orchestrator {
	if o.err != nil {
		return o
	}

	// Agent requires model component
	if o.modelComponents == nil {
		o.err = fmt.Errorf("agent requires model component; call WithModel() first")
		return o
	}

	o.agentComponent, o.mcpComponents, o.err = InitializeAgentComponent(o.ctx, o.cfg, o.modelComponents.LLM)
	return o
}

// WithSession initializes session components
func (o *Orchestrator) WithSession() *Orchestrator {
	if o.err != nil {
		return o
	}

	// Session requires agent and display components
	if o.agentComponent == nil {
		o.err = fmt.Errorf("session requires agent component; call WithAgent() first")
		return o
	}
	if o.displayComponents == nil {
		o.err = fmt.Errorf("session requires display component; call WithDisplay() first")
		return o
	}

	o.sessionComponents, o.err = InitializeSessionComponents(o.ctx, o.cfg, o.agentComponent, o.displayComponents.BannerRenderer)
	return o
}

// Build returns the orchestrated components or an error if any step failed
func (o *Orchestrator) Build() (*Components, error) {
	if o.err != nil {
		return nil, o.err
	}

	// Verify all required components are present
	if o.displayComponents == nil {
		return nil, fmt.Errorf("display components not initialized; call WithDisplay()")
	}
	if o.modelComponents == nil {
		return nil, fmt.Errorf("model components not initialized; call WithModel()")
	}
	if o.agentComponent == nil {
		return nil, fmt.Errorf("agent component not initialized; call WithAgent()")
	}
	if o.sessionComponents == nil {
		return nil, fmt.Errorf("session components not initialized; call WithSession()")
	}

	return &Components{
		Display: o.displayComponents,
		Model:   o.modelComponents,
		Agent:   o.agentComponent,
		MCP:     o.mcpComponents,
		Session: o.sessionComponents,
	}, nil
}

// Components holds all orchestrated application components
type Components struct {
	Display *DisplayComponents
	Model   *ModelComponents
	Agent   agent.Agent
	MCP     *MCPComponents
	Session *SessionComponents
}

// DisplayRenderer returns the display renderer component
func (c *Components) DisplayRenderer() *DisplayComponents {
	return c.Display
}

// ModelRegistry returns the model registry
func (c *Components) ModelRegistry() *ModelComponents {
	return c.Model
}

// AgentComponent returns the agent component
func (c *Components) AgentComponent() agent.Agent {
	return c.Agent
}

// MCPManager returns the MCP components
func (c *Components) MCPManager() *MCPComponents {
	return c.MCP
}

// SessionManager returns the session components
func (c *Components) SessionManager() *SessionComponents {
	return c.Session
}
