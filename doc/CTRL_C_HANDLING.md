# Ctrl+C and Signal Handling Improvements

## Overview

The CLI now has comprehensive and graceful Ctrl+C handling throughout the application. Users can safely interrupt operations at any point without causing issues.

## Signal Handling Features

### 1. **At Prompt (Waiting for Input)**
When the user presses Ctrl+C while the CLI is waiting for a command:
- Immediately displays goodbye message
- Cleanly exits the program
- No hanging processes or orphaned resources

```
❯ [User presses Ctrl+C]

Goodbye! Happy coding! 👋
[Program exits]
```

### 2. **During Agent Execution**
When the user presses Ctrl+C while the agent is thinking/executing:
- **First Ctrl+C**: Cancels the current agent task and returns to prompt
  - Agent operation is gracefully interrupted
  - Results are not saved if incomplete
  - User can continue with new commands
  
```
⠙ Agent is thinking [task running...]
[User presses Ctrl+C]

⚠️  Interrupted by user (Ctrl+C)
Cancelling current operation...

[Agent operation stops]
⚠️  Task cancelled by user
❯ [Ready for next command]
```

- **Second Ctrl+C** (if pressed while exiting): Force exit
  - If user presses Ctrl+C again within a few seconds
  - Program force-exits immediately with exit code 130
  - Ensures no hanging processes

```
⠙ Agent is thinking...
[User presses Ctrl+C]
⚠️  Interrupted by user (Ctrl+C)
[User presses Ctrl+C again]

⚠️  Ctrl+C pressed again - forcing exit
[Program exits with code 130]
```

### 3. **History Preservation**
- Commands are saved to `~/.code_agent_history` before execution
- Interrupted commands are still in history for reuse
- No data loss on interruption

## Implementation Details

### Signal Handler
```go
// Handles OS signals (SIGINT, SIGTERM)
go func() {
    for sig := range sigChan {
        ctrlCCount++
        if sig == syscall.SIGINT {
            if ctrlCCount == 1 {
                fmt.Println("\n\n⚠️  Interrupted by user (Ctrl+C)")
                fmt.Println("Cancelling current operation...")
            } else {
                fmt.Println("\n\n⚠️  Ctrl+C pressed again - forcing exit")
                os.Exit(130) // Standard exit code for SIGINT
            }
        }
        cancel()
    }
}()
```

### Context Cancellation
- Main context is created with `context.WithCancel()`
- Signal handler calls `cancel()` to interrupt all operations
- All goroutines check context with `ctx.Done()`

### Readline Integration
- Readline has its own interrupt handling
- When Ctrl+C at prompt, readline returns `ErrInterrupt`
- We catch this and exit gracefully

### Agent Loop
```go
agentLoop:
for event, err := range agentRunner.Run(ctx, userID, sessionName, userMsg, ...) {
    // Check if context was cancelled (Ctrl+C)
    select {
    case <-ctx.Done():
        spinner.StopWithError("Task interrupted")
        fmt.Printf("\n%s\n", renderer.Yellow("⚠️  Task cancelled by user"))
        hasError = true
        break agentLoop
    default:
    }
    // ... process event
}
```

## User Behavior Guide

### Scenario 1: Exit While Waiting for Input
```bash
$ ./code-agent
❯ [Press Ctrl+C]

Goodbye! Happy coding! 👋
$ # Program exits cleanly
```

### Scenario 2: Cancel Running Agent Task
```bash
$ ./code-agent
❯ List all files
⠙ Agent is thinking [running...]
   [Press Ctrl+C]

⚠️  Interrupted by user (Ctrl+C)
Cancelling current operation...

⚠️  Task cancelled by user
❯ # Ready for next command
```

### Scenario 3: Force Exit if Needed
```bash
$ ./code-agent
❯ [Long running task]
⠙ Agent is thinking...
[Press Ctrl+C] - cancels operation
⚠️  Interrupted by user (Ctrl+C)
[Press Ctrl+C again quickly]

⚠️  Ctrl+C pressed again - forcing exit
$ # Program force-exits
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Normal exit (user typed `/exit` or `/quit`) |
| 1 | Error during initialization |
| 130 | SIGINT (Ctrl+C forced exit) |

## Implementation Benefits

1. **Graceful Shutdown**: All resources cleaned up properly
2. **No Hanging Processes**: Signal handler ensures complete exit
3. **Two-Step Safety**: First Ctrl+C cancels, second forces exit
4. **Context Propagation**: Cancellation signal reaches all goroutines
5. **Data Preservation**: Commands saved before execution
6. **Clear Feedback**: User knows what happened

## Technical Details

### Signal Handling Flow
1. OS sends SIGINT signal (Ctrl+C)
2. Signal handler goroutine receives it
3. Counter incremented
4. Context cancelled via `cancel()`
5. All goroutines check `ctx.Done()` and exit
6. Main loop detects context cancellation and exits

### Context Cancellation
- Used for both graceful shutdown and interrupt handling
- Propagates through agent runner to all operations
- Checked at strategic points in loops

### Readline Safety
- Readline also respects context cancellation
- Has built-in interrupt handling
- We handle `readline.ErrInterrupt` appropriately

## Testing Ctrl+C Behavior

### Test 1: Exit at Prompt
```bash
./code-agent
# Wait for prompt
# Press Ctrl+C
# Expected: Goodbye message and exit
```

### Test 2: Cancel During Agent
```bash
./code-agent
# Type a command that takes time (e.g., "list all files in ./")
# Press Ctrl+C immediately
# Expected: Task cancelled message, ready for next command
```

### Test 3: Normal Exit
```bash
./code-agent
# Type /exit or /quit
# Expected: Goodbye message and exit
```

## Future Enhancements

Possible improvements:
1. Save incomplete agent results to history
2. Configurable double-Ctrl+C timeout
3. Agent graceful shutdown hooks
4. Logging of interrupted operations
5. Option to resume interrupted tasks
