package commands

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"adk-code/internal/display"
	"adk-code/internal/tracking"
	"adk-code/pkg/models"
)

// REPLCommand defines the interface for REPL commands
type REPLCommand interface {
	// Name returns the command name (e.g., "help", "tools", "models")
	Name() string

	// Description returns brief help text
	Description() string

	// Execute runs the command with given arguments
	// ctx is the execution context, args are command-specific arguments
	Execute(ctx context.Context, args []string) error
}

// CommandRegistry manages available REPL commands
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]REPLCommand
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]REPLCommand),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd REPLCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name
// Returns nil if command not found
func (r *CommandRegistry) Get(name string) REPLCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commands[name]
}

// List returns all registered command names
func (r *CommandRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// Has checks if a command is registered
func (r *CommandRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.commands[name]
	return ok
}

// PromptCommand implements REPLCommand for /prompt
type PromptCommand struct {
	renderer *display.Renderer
}

// NewPromptCommand creates a new prompt command
func NewPromptCommand(renderer *display.Renderer) *PromptCommand {
	return &PromptCommand{
		renderer: renderer,
	}
}

// Name returns the command name
func (c *PromptCommand) Name() string {
	return "prompt"
}

// Description returns command help text
func (c *PromptCommand) Description() string {
	return "Display the system prompt (XML-structured)"
}

// Execute runs the prompt command
func (c *PromptCommand) Execute(ctx context.Context, args []string) error {
	// Original implementation from handlePromptCommand
	return c.execute()
}

// execute is the actual implementation
func (c *PromptCommand) execute() error {
	// Use the original handler logic
	handlePromptCommand(c.renderer)
	return nil
}

// HelpCommand implements REPLCommand for /help
type HelpCommand struct {
	renderer *display.Renderer
}

// NewHelpCommand creates a new help command
func NewHelpCommand(renderer *display.Renderer) *HelpCommand {
	return &HelpCommand{
		renderer: renderer,
	}
}

// Name returns the command name
func (c *HelpCommand) Name() string {
	return "help"
}

// Description returns command help text
func (c *HelpCommand) Description() string {
	return "Display help information about available commands"
}

// Execute runs the help command
func (c *HelpCommand) Execute(ctx context.Context, args []string) error {
	handleHelpCommand(c.renderer)
	return nil
}

// ToolsCommand implements REPLCommand for /tools
type ToolsCommand struct {
	renderer *display.Renderer
}

// NewToolsCommand creates a new tools command
func NewToolsCommand(renderer *display.Renderer) *ToolsCommand {
	return &ToolsCommand{
		renderer: renderer,
	}
}

// Name returns the command name
func (c *ToolsCommand) Name() string {
	return "tools"
}

// Description returns command help text
func (c *ToolsCommand) Description() string {
	return "List all available tools"
}

// Execute runs the tools command
func (c *ToolsCommand) Execute(ctx context.Context, args []string) error {
	handleToolsCommand(c.renderer)
	return nil
}

// ModelsCommand implements REPLCommand for /models
type ModelsCommand struct {
	renderer *display.Renderer
	registry *models.Registry
}

// NewModelsCommand creates a new models command
func NewModelsCommand(renderer *display.Renderer, registry *models.Registry) *ModelsCommand {
	return &ModelsCommand{
		renderer: renderer,
		registry: registry,
	}
}

// Name returns the command name
func (c *ModelsCommand) Name() string {
	return "models"
}

// Description returns command help text
func (c *ModelsCommand) Description() string {
	return "List all available AI models"
}

// Execute runs the models command
func (c *ModelsCommand) Execute(ctx context.Context, args []string) error {
	handleModelsCommand(c.renderer, c.registry)
	return nil
}

// CurrentModelCommand implements REPLCommand for /current-model
type CurrentModelCommand struct {
	renderer     *display.Renderer
	currentModel models.Config
}

// NewCurrentModelCommand creates a new current model command
func NewCurrentModelCommand(renderer *display.Renderer, currentModel models.Config) *CurrentModelCommand {
	return &CurrentModelCommand{
		renderer:     renderer,
		currentModel: currentModel,
	}
}

// Name returns the command name
func (c *CurrentModelCommand) Name() string {
	return "current-model"
}

// Description returns command help text
func (c *CurrentModelCommand) Description() string {
	return "Show details about the current AI model"
}

// Execute runs the current model command
func (c *CurrentModelCommand) Execute(ctx context.Context, args []string) error {
	handleCurrentModelCommand(c.renderer, c.currentModel)
	return nil
}

// ProvidersCommand implements REPLCommand for /providers
type ProvidersCommand struct {
	renderer *display.Renderer
	registry *models.Registry
}

// NewProvidersCommand creates a new providers command
func NewProvidersCommand(renderer *display.Renderer, registry *models.Registry) *ProvidersCommand {
	return &ProvidersCommand{
		renderer: renderer,
		registry: registry,
	}
}

// Name returns the command name
func (c *ProvidersCommand) Name() string {
	return "providers"
}

// Description returns command help text
func (c *ProvidersCommand) Description() string {
	return "Show available providers and their models"
}

// Execute runs the providers command
func (c *ProvidersCommand) Execute(ctx context.Context, args []string) error {
	handleProvidersCommand(ctx, c.renderer, c.registry)
	return nil
}

// TokensCommand implements REPLCommand for /tokens
type TokensCommand struct {
	sessionTokens *tracking.SessionTokens
}

// NewTokensCommand creates a new tokens command
func NewTokensCommand(sessionTokens *tracking.SessionTokens) *TokensCommand {
	return &TokensCommand{
		sessionTokens: sessionTokens,
	}
}

// Name returns the command name
func (c *TokensCommand) Name() string {
	return "tokens"
}

// Description returns command help text
func (c *TokensCommand) Description() string {
	return "Display token usage statistics"
}

// Execute runs the tokens command
func (c *TokensCommand) Execute(ctx context.Context, args []string) error {
	handleTokensCommand(c.sessionTokens)
	return nil
}

// SetModelCommand implements REPLCommand for /set-model
type SetModelCommand struct {
	renderer *display.Renderer
	registry *models.Registry
}

// NewSetModelCommand creates a new set model command
func NewSetModelCommand(renderer *display.Renderer, registry *models.Registry) *SetModelCommand {
	return &SetModelCommand{
		renderer: renderer,
		registry: registry,
	}
}

// Name returns the command name
func (c *SetModelCommand) Name() string {
	return "set-model"
}

// Description returns command help text
func (c *SetModelCommand) Description() string {
	return "Validate and plan to switch to a different model"
}

// Execute runs the set model command
func (c *SetModelCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify a model using provider/model syntax")
	}

	modelSpec := strings.Join(args, " ")
	HandleSetModel(c.renderer, c.registry, modelSpec)
	return nil
}

// NewDefaultCommandRegistry creates a command registry with all standard REPL commands
func NewDefaultCommandRegistry(
	renderer *display.Renderer,
	modelRegistry *models.Registry,
	currentModel models.Config,
	sessionTokens *tracking.SessionTokens,
) *CommandRegistry {
	registry := NewCommandRegistry()

	// Register all built-in commands
	registry.Register(NewHelpCommand(renderer))
	registry.Register(NewPromptCommand(renderer))
	registry.Register(NewToolsCommand(renderer))
	registry.Register(NewModelsCommand(renderer, modelRegistry))
	registry.Register(NewCurrentModelCommand(renderer, currentModel))
	registry.Register(NewProvidersCommand(renderer, modelRegistry))
	registry.Register(NewTokensCommand(sessionTokens))
	registry.Register(NewSetModelCommand(renderer, modelRegistry))

	return registry
}
