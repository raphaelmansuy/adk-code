# Model Selection and Display Improvements

## Overview

The code agent now supports flexible model selection with enhanced display capabilities. Users can easily switch between different AI models (Gemini and Vertex AI) with a new registry system and improved CLI interface.

## New Features

### 1. Model Registry System (`models.go`)

A centralized registry that manages available models and their configurations:

- **ModelConfig**: Stores comprehensive model metadata
  - Model ID and display name
  - Backend type (Gemini API or Vertex AI)
  - Context window size
  - Capabilities (vision support, tool use, long context)
  - Cost tier classification
  - Recommended use cases
  - Default flag

- **ModelRegistry**: Manages and queries available models
  - `RegisterModel()`: Add new models to registry
  - `GetModel()`: Retrieve specific model by ID
  - `GetModelByName()`: Retrieve by display name (case-insensitive)
  - `GetDefaultModel()`: Get the default model
  - `ListModels()`: Get all registered models
  - `ListModelsByBackend()`: Filter models by backend
  - `ResolveModel()`: Smart model selection based on user input and context

### 2. Available Models

The registry comes pre-configured with popular models:

#### Gemini API Models:
- `gemini-2.5-flash` (✓ Default)
  - Fast, affordable, multimodal
  - Context: 1M tokens
  - Cost Tier: Economy
  - Recommended for: Coding, analysis, rapid iteration

- `gemini-2.0-flash`
  - Previous generation fast model
  - Context: 1M tokens
  - Cost Tier: Economy
  - Recommended for: Coding, prototyping

- `gemini-1.5-flash`
  - Earlier flash model with large context
  - Context: 1M tokens
  - Cost Tier: Economy
  - Recommended for: Coding, document processing

- `gemini-1.5-pro`
  - Advanced reasoning model
  - Context: 2M tokens
  - Cost Tier: Premium
  - Recommended for: Complex reasoning, analysis, creative

#### Vertex AI Models:
- `gemini-2.5-flash-vertex`: Gemini 2.5 Flash via Vertex AI
- `gemini-1.5-pro-vertex`: Gemini 1.5 Pro via Vertex AI

### 3. CLI Enhancements

#### New `--model` Flag
```bash
# Select a specific model
./code-agent --model gemini-1.5-pro

# Select Vertex AI variant
./code-agent --model gemini-2.5-flash-vertex --backend vertexai

# Full example
./code-agent \
  --model gemini-1.5-pro \
  --backend gemini \
  --session my-session \
  --working-directory /path/to/project
```

#### Help Output
```bash
./code-agent --help
```
Now shows:
```
  -model string
        Model to use (e.g., gemini-2.5-flash, gemini-1.5-pro). Use '/models' command to list available models.
```

### 4. New Interactive Commands

#### `/models` - List All Available Models
Displays all registered models grouped by backend with:
- Model name and cost tier icon
- Description
- Context window size
- Tool use and vision support
- Current selection indicator (✓)

Example output:
```
════════════════════════════════════════════════════════════════
                    Available AI Models
════════════════════════════════════════════════════════════════

🔷 Gemini API Models:
   ✓ 💵 Gemini 2.5 Flash - Fast, affordable multimodal model. Best for real-time applications.
      Context: 1000000 tokens | Tools: true | Vision: true
   
   ○ 💵 Gemini 1.5 Pro - Advanced reasoning model. Best for complex tasks.
      Context: 2000000 tokens | Tools: true | Vision: true

🔶 Vertex AI Models:
   ○ 💵 Gemini 2.5 Flash (Vertex AI) - Fast, affordable model via Vertex AI...
      Context: 1000000 tokens | Tools: true | Vision: true

Use --model flag to select a model (e.g., --model gemini-1.5-pro)
Use /current-model command to see details about the active model
```

#### `/current-model` - Show Current Model Details
Displays detailed information about the actively selected model:
- Model name and backend type
- Detailed description
- Full capability matrix (vision, tools, long context)
- Context window and cost tier
- Recommended use cases

Example output:
```
════════════════════════════════════════════════════════════════
                    Current Model Information
════════════════════════════════════════════════════════════════

Model: 🔷 Gemini 2.5 Flash (gemini)

Description:
  Fast, affordable multimodal model. Best for real-time applications.

Capabilities:
  ✓ Vision/Image Processing
  ✓ Tool/Function Calling
  ✓ Long Context Window (1M+ tokens)

Technical Details:
  Context Window: 1000000 tokens
  Cost Tier: economy

Recommended For:
  • coding
  • analysis
  • rapid iteration

Tip: Use --model flag to switch models when starting the agent
```

### 5. Smart Model Resolution

The system uses a priority-based resolution strategy:

1. **Explicit model ID** (`--model gemini-1.5-pro`) - Highest priority
2. **Explicit backend** (`--backend vertexai`) - Use default for that backend
3. **Environment-based auto-detect** - Based on env vars
4. **Global default** - Fallback to `gemini-2.5-flash`

Example:
```bash
# Uses gemini-1.5-pro regardless of backend
./code-agent --model gemini-1.5-pro

# Uses Vertex AI's default model (gemini-2.5-flash-vertex)
./code-agent --backend vertexai

# Uses Gemini 2.5 Flash (default)
./code-agent
```

### 6. Enhanced Banner Display

The welcome banner now shows the selected model with proper context:

```
════════════════════════════════════════════════════════════════
  code_agent v1.0.0
  Gemini 2.5 Flash
  ~/projects/my-project
════════════════════════════════════════════════════════════════
```

For Vertex AI:
```
════════════════════════════════════════════════════════════════
  code_agent v1.0.0
  Gemini 2.5 Flash (Vertex AI)
  ~/projects/my-project
════════════════════════════════════════════════════════════════
```

## Implementation Details

### Files Modified

1. **`models.go`** (NEW)
   - ModelRegistry implementation
   - Model resolution logic
   - Helper functions

2. **`models_test.go`** (NEW)
   - Comprehensive test coverage for registry
   - Tests for model resolution
   - Tests for backend filtering

3. **`cli.go`**
   - Added `--model` flag to CLIConfig
   - Updated ParseCLIFlags() to parse model selection
   - Added `/models` and `/current-model` command handlers
   - Enhanced help text with model commands

4. **`main.go`**
   - Integrated model registry creation
   - Updated model resolution logic
   - Pass model info to command handlers

### Code Examples

#### Creating and Using the Registry

```go
// Create registry with default models
registry := NewModelRegistry()

// Get a specific model
model, err := registry.GetModel("gemini-1.5-pro")

// List all Gemini models
geminiBakcend := registry.ListModelsByBackend("gemini")

// Resolve model based on user input
selectedModel, err := registry.ResolveModel(cliConfig.Model, cliConfig.Backend)
```

#### Extracting Model IDs for APIs

```go
// Convert "-vertex" suffix variants to actual API model names
actualModelID := ExtractModelIDFromGemini("gemini-2.5-flash-vertex")
// Returns: "gemini-2.5-flash"
```

## Usage Examples

### Example 1: Use Advanced Model for Complex Task
```bash
./code-agent --model gemini-1.5-pro
# Then inside the agent:
# > Analyze the code and suggest architectural improvements
```

### Example 2: Use Vertex AI for Enterprise Deployment
```bash
./code-agent \
  --model gemini-2.5-flash-vertex \
  --project my-gcp-project \
  --location us-central1 \
  --session enterprise-coding
```

### Example 3: Explore Available Models
```bash
./code-agent
# Then inside:
# > /models
# Shows all available models
# > /current-model
# Shows detailed info about active model
```

### Example 4: Add New Model to Registry

In `models.go`, add to the registry initialization:

```go
registry.RegisterModel(ModelConfig{
    ID:          "my-custom-model",
    Name:        "My Custom Model",
    DisplayName: "My Custom Model",
    Backend:     "gemini",
    ContextWindow: 1000000,
    Capabilities: ModelCapabilities{
        VisionSupport:     true,
        ToolUseSupport:    true,
        LongContextWindow: true,
        CostTier:          "economy",
    },
    Description:    "My custom model description",
    RecommendedFor: []string{"use-case-1", "use-case-2"},
    IsDefault:      false,
})
```

## Testing

All features are covered by unit tests:

```bash
# Run model registry tests
cd code_agent
go test -v -run TestModel

# Run all tests
make test

# Run full quality checks
make check
```

Test coverage includes:
- Model retrieval and registration
- Model resolution with different input combinations
- Backend filtering
- Model ID extraction
- Default model selection

## Benefits

1. **Flexibility**: Easy to switch between models without recompilation
2. **User Awareness**: Clear display of current model and capabilities
3. **Extensibility**: Simple to add new models to the registry
4. **Cost Awareness**: Cost tier indicators help users choose appropriately
5. **Discoverability**: Built-in commands make available models obvious

## Future Enhancements

Potential improvements:
- Model performance benchmarks in registry
- Cost estimation per request based on model
- Save preferred model per session
- Model compatibility checking for specific tasks
- Streaming model evaluation/comparison
- Custom model registry entries from config file
