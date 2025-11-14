package execution

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"time"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeExecutor    PluginType = "executor"
	PluginTypeStrategy    PluginType = "strategy"
	PluginTypeCredential  PluginType = "credential"
	PluginTypeTransport   PluginType = "transport"
)

// PluginMetadata contains metadata about a plugin
type PluginMetadata struct {
	// Name is the plugin name
	Name string

	// Version is the plugin version
	Version string

	// Description is the plugin description
	Description string

	// Type is the plugin type
	Type PluginType

	// Author is the plugin author
	Author string

	// License is the plugin license
	License string

	// LoadedAt is when the plugin was loaded
	LoadedAt time.Time

	// FilePath is the path to the plugin file
	FilePath string
}

// PluginExecutor is the interface that executor plugins must implement
type PluginExecutor interface {
	// Execute runs the executor plugin with generic context and result types
	Execute(ctx interface{}) (interface{}, error)

	// GetMetadata returns plugin metadata
	GetMetadata() *PluginMetadata

	// Validate validates the plugin is functional
	Validate() error
}

// PluginRegistry manages loaded plugins
type PluginRegistry struct {
	// executors maps executor names to loaded plugins
	executors map[string]PluginExecutor

	// metadata maps plugin names to metadata
	metadata map[string]*PluginMetadata

	// pluginPaths are paths to search for plugins
	pluginPaths []string
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		executors:   make(map[string]PluginExecutor),
		metadata:    make(map[string]*PluginMetadata),
		pluginPaths: []string{},
	}
}

// AddPluginPath adds a path to search for plugins
func (pr *PluginRegistry) AddPluginPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("invalid plugin path %q: %w", path, err)
	}

	pr.pluginPaths = append(pr.pluginPaths, path)
	return nil
}

// LoadPlugin loads a plugin from a file
func (pr *PluginRegistry) LoadPlugin(filePath string) (*PluginMetadata, error) {
	// Load the plugin
	p, err := plugin.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// Look up the New symbol
	newSym, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("plugin must export a 'New' function: %w", err)
	}

	// Create plugin instance
	newFunc, ok := newSym.(func() (PluginExecutor, error))
	if !ok {
		return nil, fmt.Errorf("New must have signature func() (PluginExecutor, error)")
	}

	executor, err := newFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance: %w", err)
	}

	// Get metadata
	meta := executor.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("plugin must return metadata")
	}

	meta.FilePath = filePath
	meta.LoadedAt = time.Now()

	// Validate the plugin
	if err := executor.Validate(); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// Register the plugin
	pr.executors[meta.Name] = executor
	pr.metadata[meta.Name] = meta

	return meta, nil
}

// LoadPluginsFromPath loads all plugins from a directory
func (pr *PluginRegistry) LoadPluginsFromPath(dirPath string) ([]*PluginMetadata, error) {
	if _, err := os.Stat(dirPath); err != nil {
		return nil, fmt.Errorf("invalid plugin directory: %w", err)
	}

	var loaded []*PluginMetadata

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	// Load each plugin file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .so files (Unix) or .dll files (Windows)
		name := entry.Name()
		if !isPluginFile(name) {
			continue
		}

		filePath := filepath.Join(dirPath, name)
		meta, err := pr.LoadPlugin(filePath)
		if err != nil {
			// Log error but continue loading other plugins
			continue
		}

		loaded = append(loaded, meta)
	}

	return loaded, nil
}

// GetExecutor gets an executor plugin by name
func (pr *PluginRegistry) GetExecutor(name string) (PluginExecutor, error) {
	executor, exists := pr.executors[name]
	if !exists {
		return nil, fmt.Errorf("executor plugin %q not found", name)
	}

	return executor, nil
}

// GetMetadata gets plugin metadata by name
func (pr *PluginRegistry) GetMetadata(name string) (*PluginMetadata, error) {
	meta, exists := pr.metadata[name]
	if !exists {
		return nil, fmt.Errorf("metadata for plugin %q not found", name)
	}

	return meta, nil
}

// ListPlugins lists all loaded plugins
func (pr *PluginRegistry) ListPlugins() []*PluginMetadata {
	var plugins []*PluginMetadata
	for _, meta := range pr.metadata {
		plugins = append(plugins, meta)
	}
	return plugins
}

// UnloadPlugin unloads a plugin
func (pr *PluginRegistry) UnloadPlugin(name string) error {
	if _, exists := pr.executors[name]; !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	delete(pr.executors, name)
	delete(pr.metadata, name)

	return nil
}

// PluginConfig represents plugin configuration
type PluginConfig struct {
	// Name is the plugin name
	Name string

	// Path is the plugin file path
	Path string

	// Enabled indicates if the plugin is enabled
	Enabled bool

	// Settings are plugin-specific settings
	Settings map[string]interface{}
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	registry *PluginRegistry
	configs  map[string]*PluginConfig
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		registry: NewPluginRegistry(),
		configs:  make(map[string]*PluginConfig),
	}
}

// AddPluginConfig adds a plugin configuration
func (pm *PluginManager) AddPluginConfig(config *PluginConfig) error {
	if config.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if config.Path == "" {
		return fmt.Errorf("plugin path is required")
	}

	pm.configs[config.Name] = config
	return nil
}

// LoadConfiguredPlugins loads plugins based on configuration
func (pm *PluginManager) LoadConfiguredPlugins() (map[string]*PluginMetadata, error) {
	loaded := make(map[string]*PluginMetadata)

	for name, config := range pm.configs {
		if !config.Enabled {
			continue
		}

		meta, err := pm.registry.LoadPlugin(config.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin %q: %w", name, err)
		}

		loaded[name] = meta
	}

	return loaded, nil
}

// GetRegistry returns the underlying plugin registry
func (pm *PluginManager) GetRegistry() *PluginRegistry {
	return pm.registry
}

// isPluginFile checks if a file is a plugin file
func isPluginFile(name string) bool {
	// On Unix, plugins are .so files
	// On Windows, plugins are .dll files
	ext := filepath.Ext(name)
	return ext == ".so" || ext == ".dll" || ext == ".dylib"
}

// PluginValidator validates plugin compatibility
type PluginValidator struct {
	minVersion string
	maxVersion string
}

// NewPluginValidator creates a new plugin validator
func NewPluginValidator(minVersion, maxVersion string) *PluginValidator {
	return &PluginValidator{
		minVersion: minVersion,
		maxVersion: maxVersion,
	}
}

// ValidatePlugin validates a plugin meets requirements
func (pv *PluginValidator) ValidatePlugin(meta *PluginMetadata) error {
	if meta == nil {
		return fmt.Errorf("plugin metadata is nil")
	}

	if meta.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if meta.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	// In a real implementation, this would compare versions
	// For now, just verify the plugin loaded successfully
	return nil
}

// PluginEvent represents an event in plugin lifecycle
type PluginEvent struct {
	// Type is the event type
	Type string

	// PluginName is the name of the affected plugin
	PluginName string

	// Timestamp is when the event occurred
	Timestamp time.Time

	// Message is the event message
	Message string
}

// PluginEventListener is called when plugin events occur
type PluginEventListener func(event *PluginEvent)

// PluginEventBus manages plugin events
type PluginEventBus struct {
	listeners []PluginEventListener
}

// NewPluginEventBus creates a new plugin event bus
func NewPluginEventBus() *PluginEventBus {
	return &PluginEventBus{
		listeners: []PluginEventListener{},
	}
}

// Subscribe adds an event listener
func (peb *PluginEventBus) Subscribe(listener PluginEventListener) {
	peb.listeners = append(peb.listeners, listener)
}

// Emit emits an event to all listeners
func (peb *PluginEventBus) Emit(event *PluginEvent) {
	for _, listener := range peb.listeners {
		listener(event)
	}
}
