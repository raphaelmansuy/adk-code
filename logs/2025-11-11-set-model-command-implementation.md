# Set Model Command Implementation

**Date**: November 11, 2025  
**Task**: Ensure we have a command within CLI to set the current model

## Summary

Successfully implemented the `/set-model` command that allows users to validate and plan switching to different AI models within the REPL.

## Implementation Details

### Command Handler Logic

The `/set-model` command is implemented in two parts:

1. **Command Recognition** - Modified `handleBuiltinCommand()` to detect `/set-model <spec>` input in the default case
2. **Validation & Display** - New `handleSetModel()` function validates the model and provides helpful guidance

### Smart Model Suggestion

The implementation intelligently suggests the best command format:

- **Full syntax provided** (`gemini/2.5-flash`) → Echo back as-is
- **Shorthand provided** (`gemini/flash`) → Echo back as-is  
- **Model ID only** (`gemini-2.0-flash`) → Extract shorthand to suggest `--model gemini/2.0-flash`

This is handled by the `extractShorthandFromModelID()` helper function that strips the provider prefix.

## Changes Made

### 1. **Added `/set-model` Command Handler** (`cli.go`)

- Modified `handleBuiltinCommand()` to recognize `/set-model <provider/model>` input
- Added check in the `default` case to parse commands with arguments (like `/set-model`)
- Command extracts the model specification and passes it to the new `handleSetModel()` handler

### 2. **Implemented `handleSetModel()` Function** (`cli.go`)

The function:

- Validates the model syntax using existing `ParseProviderModelSyntax()`
- Resolves the model using the registry's `ResolveFromProviderSyntax()` method
- Displays detailed information about the selected model including:
  - Model name and backend (Gemini or Vertex AI)
  - Context window size
  - Cost tier information
- Provides helpful guidance on how to restart with the new model

### 3. **Added `extractShorthandFromModelID()` Helper** (`cli.go`)

Intelligently extracts shorthands from full model IDs:

- `gemini-2.5-flash` → `2.5-flash`
- `gemini-1.5-pro` → `1.5-pro`
- `gemini-2.0-flash` → `2.0-flash`

### 4. **Updated Help Message** (`cli.go`)

Added the new command to the built-in commands list:

```text
• /set-model <provider/model> - Validate and plan to switch models
```

## How It Works

Users can now type within the REPL:

```bash
/set-model gemini/2.5-flash
/set-model gemini/flash
/set-model vertexai/1.5-pro
/set-model gemini-1.5-pro
```

The command will:

1. Validate that the model exists
2. Display information about the selected model
3. Inform users that the model can only be switched at startup
4. Show the exact command to restart with the new model

### Example Output

```text
✓ Model validation successful!

You selected: Gemini 2.5 Flash (gemini)
Context window: 1000000 tokens
Cost tier: economy

ℹ️  Note:
The model can only be switched at startup. To actually use this model, exit the agent and restart with:
  ./code-agent --model gemini/2.5-flash
```

## Error Handling

If an invalid model is specified:

- User gets a clear error message
- List of available providers and models is displayed
- Helps users discover valid model options

Example error output:

```text
Model not found: model "invalid-model" not found for provider "gemini"

Available models:

Gemini:
  • gemini/2.5-flash
  • gemini/2.0-flash
  • gemini-1.5-flash
  • gemini/1.5-pro

Vertex AI:
  • vertexai/2.5-flash
  • vertexai/2.0-flash
  • vertexai/1.5-flash
  • vertexai/1.5-pro
```

## Design Decisions

1. **Read-only validation**: The command validates models without actually switching them because switching LLM models requires reinitializing the genai client and the agent's model instance.

2. **Clear guidance**: Instead of silently failing or showing confusing errors, the command clearly explains that models must be set at startup and shows the exact command needed.

3. **Reuses existing infrastructure**: Leverages existing `ParseProviderModelSyntax()`, `ResolveFromProviderSyntax()`, and the `ModelRegistry` system for consistency.

## Testing

- ✅ Code compiles without errors
- ✅ `go fmt` and `go vet` pass
- ✅ No linting issues
- ✅ Integration with existing command handling system verified

## Future Enhancements

Potential future improvements:

1. Store user's preferred model selection across sessions
2. Add model switching with agent reinitialization (requires architectural changes)
3. Show model pricing estimates for selected models
