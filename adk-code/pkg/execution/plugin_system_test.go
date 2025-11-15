package execution

import (
	"testing"
	"time"
)

// TestPluginMetadata tests plugin metadata
func TestPluginMetadata(t *testing.T) {
	meta := &PluginMetadata{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Type:        PluginTypeExecutor,
		Author:      "Test Author",
		License:     "MIT",
		LoadedAt:    time.Now(),
		FilePath:    "/path/to/plugin.so",
	}

	if meta.Name != "test-plugin" {
		t.Fatalf("Expected name 'test-plugin', got %q", meta.Name)
	}

	if meta.Type != PluginTypeExecutor {
		t.Fatalf("Expected type PluginTypeExecutor, got %v", meta.Type)
	}
}

// TestPluginType tests plugin type constants
func TestPluginType(t *testing.T) {
	types := []PluginType{
		PluginTypeExecutor,
		PluginTypeStrategy,
		PluginTypeCredential,
		PluginTypeTransport,
	}

	if len(types) != 4 {
		t.Fatalf("Expected 4 plugin types, got %d", len(types))
	}

	if types[0] != PluginTypeExecutor {
		t.Fatal("Expected PluginTypeExecutor")
	}
}

// TestPluginRegistry tests basic registry operations
func TestPluginRegistry(t *testing.T) {
	registry := NewPluginRegistry()

	if registry == nil {
		t.Fatal("Failed to create registry")
	}

	plugins := registry.ListPlugins()
	if len(plugins) != 0 {
		t.Fatalf("Expected 0 plugins initially, got %d", len(plugins))
	}
}

// TestPluginRegistryAddPath tests adding plugin paths
func TestPluginRegistryAddPath(t *testing.T) {
	registry := NewPluginRegistry()

	// Try to add a non-existent path
	err := registry.AddPluginPath("/non/existent/path")
	if err == nil {
		t.Fatal("Expected error for non-existent path")
	}
}

// TestPluginConfig tests plugin configuration
func TestPluginConfig(t *testing.T) {
	config := &PluginConfig{
		Name:    "test-plugin",
		Path:    "/path/to/plugin.so",
		Enabled: true,
		Settings: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
	}

	if config.Name != "test-plugin" {
		t.Fatalf("Expected name 'test-plugin', got %q", config.Name)
	}

	if !config.Enabled {
		t.Fatal("Expected plugin to be enabled")
	}

	if timeout, ok := config.Settings["timeout"].(int); ok && timeout != 30 {
		t.Fatalf("Expected timeout 30, got %d", timeout)
	}
}

// TestPluginManager tests plugin manager creation
func TestPluginManager(t *testing.T) {
	manager := NewPluginManager()

	if manager == nil {
		t.Fatal("Failed to create plugin manager")
	}

	if manager.GetRegistry() == nil {
		t.Fatal("Expected registry to be initialized")
	}
}

// TestPluginManagerAddConfig tests adding plugin configuration
func TestPluginManagerAddConfig(t *testing.T) {
	manager := NewPluginManager()

	config := &PluginConfig{
		Name:    "test-plugin",
		Path:    "/path/to/plugin.so",
		Enabled: true,
	}

	err := manager.AddPluginConfig(config)
	if err != nil {
		t.Fatalf("Failed to add plugin config: %v", err)
	}
}

// TestPluginManagerAddConfigMissingName tests adding config with missing name
func TestPluginManagerAddConfigMissingName(t *testing.T) {
	manager := NewPluginManager()

	config := &PluginConfig{
		Path:    "/path/to/plugin.so",
		Enabled: true,
	}

	err := manager.AddPluginConfig(config)
	if err == nil {
		t.Fatal("Expected error for missing plugin name")
	}
}

// TestPluginManagerAddConfigMissingPath tests adding config with missing path
func TestPluginManagerAddConfigMissingPath(t *testing.T) {
	manager := NewPluginManager()

	config := &PluginConfig{
		Name:    "test-plugin",
		Enabled: true,
	}

	err := manager.AddPluginConfig(config)
	if err == nil {
		t.Fatal("Expected error for missing plugin path")
	}
}

// TestPluginValidator tests plugin validation
func TestPluginValidator(t *testing.T) {
	validator := NewPluginValidator("1.0.0", "2.0.0")

	meta := &PluginMetadata{
		Name:    "test-plugin",
		Version: "1.5.0",
	}

	err := validator.ValidatePlugin(meta)
	if err != nil {
		t.Fatalf("Failed to validate plugin: %v", err)
	}
}

// TestPluginValidatorMissingName tests validating plugin without name
func TestPluginValidatorMissingName(t *testing.T) {
	validator := NewPluginValidator("1.0.0", "2.0.0")

	meta := &PluginMetadata{
		Version: "1.5.0",
	}

	err := validator.ValidatePlugin(meta)
	if err == nil {
		t.Fatal("Expected error validating plugin without name")
	}
}

// TestPluginValidatorNilMetadata tests validating nil metadata
func TestPluginValidatorNilMetadata(t *testing.T) {
	validator := NewPluginValidator("1.0.0", "2.0.0")

	err := validator.ValidatePlugin(nil)
	if err == nil {
		t.Fatal("Expected error validating nil metadata")
	}
}

// TestPluginEvent tests plugin event creation
func TestPluginEvent(t *testing.T) {
	event := &PluginEvent{
		Type:       "loaded",
		PluginName: "test-plugin",
		Timestamp:  time.Now(),
		Message:    "Plugin loaded successfully",
	}

	if event.Type != "loaded" {
		t.Fatalf("Expected type 'loaded', got %q", event.Type)
	}

	if event.PluginName != "test-plugin" {
		t.Fatalf("Expected plugin name 'test-plugin', got %q", event.PluginName)
	}
}

// TestPluginEventBus tests event bus creation
func TestPluginEventBus(t *testing.T) {
	bus := NewPluginEventBus()

	if bus == nil {
		t.Fatal("Failed to create event bus")
	}
}

// TestPluginEventBusEmit tests emitting events
func TestPluginEventBusEmit(t *testing.T) {
	bus := NewPluginEventBus()

	eventReceived := false
	bus.Subscribe(func(event *PluginEvent) {
		eventReceived = true
	})

	event := &PluginEvent{
		Type:       "loaded",
		PluginName: "test-plugin",
		Message:    "Plugin loaded",
	}

	bus.Emit(event)

	if !eventReceived {
		t.Fatal("Expected event to be received by listener")
	}
}

// TestPluginEventBusMultipleListeners tests multiple event listeners
func TestPluginEventBusMultipleListeners(t *testing.T) {
	bus := NewPluginEventBus()

	count := 0
	for i := 0; i < 3; i++ {
		bus.Subscribe(func(event *PluginEvent) {
			count++
		})
	}

	event := &PluginEvent{
		Type:       "loaded",
		PluginName: "test-plugin",
	}

	bus.Emit(event)

	if count != 3 {
		t.Fatalf("Expected 3 listener calls, got %d", count)
	}
}

// TestPluginMetadataFilePath tests plugin metadata with file path
func TestPluginMetadataFilePath(t *testing.T) {
	meta := &PluginMetadata{
		Name:     "test-plugin",
		FilePath: "/usr/lib/plugins/test-plugin.so",
		LoadedAt: time.Now(),
	}

	if meta.FilePath != "/usr/lib/plugins/test-plugin.so" {
		t.Fatalf("Expected file path '/usr/lib/plugins/test-plugin.so', got %q", meta.FilePath)
	}
}

// TestIsPluginFile tests plugin file detection
func TestIsPluginFile(t *testing.T) {
	testCases := []struct {
		filename string
		isPlugin bool
	}{
		{"plugin.so", true},
		{"plugin.dll", true},
		{"plugin.dylib", true},
		{"plugin.o", false},
		{"plugin.a", false},
		{"plugin", false},
	}

	for _, tc := range testCases {
		result := isPluginFile(tc.filename)
		if result != tc.isPlugin {
			t.Fatalf("isPluginFile(%q) = %v, expected %v", tc.filename, result, tc.isPlugin)
		}
	}
}

// TestPluginManagerRegistry tests accessing registry from manager
func TestPluginManagerRegistry(t *testing.T) {
	manager := NewPluginManager()
	registry := manager.GetRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be returned")
	}

	if len(registry.ListPlugins()) != 0 {
		t.Fatal("Expected empty plugin list initially")
	}
}
