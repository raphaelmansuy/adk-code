# Quick Reference: Cline Features for Code Agent

## One-Page Summary of High-Value Features

### 🏆 Top 5 Must-Have Features

| Feature | What it does | Why it matters | Effort |
|---------|-------------|---|--------|
| **Checkpoints** | Snapshot/restore workspace state | Enables safe experimentation and easy rollback | High |
| **Focus Chain** | Auto-summarize when context low | Critical for long-running tasks | Medium |
| **Mention System** | @file, @folder, @url context injection | Dramatically improves UX | Medium |
| **Auto-Approval** | Granular permission levels | Essential for safe autonomous operation | Low |
| **MCP Integration** | Extensible custom tools | Future-proofs architecture | High |

---

## Feature Categories Quick Reference

### 🛡️ Safety & Control
- **Auto-Approval** - Granular permissions (readFiles, editFiles, executeSafeCommands, useBrowser, useMcp)
- **YOLO Mode** - Override all approvals for power users
- **Workspace Boundaries** - Different rules for workspace vs external paths
- **Mistake Tracking** - Count consecutive mistakes and adjust behavior

### 📝 State Management
- **Task Persistence** - Full conversation history to disk
- **Checkpoints** - Shadow git repo for workspace snapshots
- **Focus Chain** - Markdown checklist progress tracking
- **History Reconstruction** - Can replay tasks from history

### 🎯 UX/Input
- **Mention System** - @file, @folder, @url, @problems, @terminal, @git-changes
- **Slash Commands** - /newtask, /deep-planning, /smol, /newrule, /reportbug
- **Multi-root Syntax** - @workspace:path/file for monorepos
- **Auto-completion** - Suggest file/folder mentions

### 🧠 Intelligence Features
- **Deep Planning Mode** - Structured investigation → questions → plan → task
- **Focus Chain** - Auto-summarize context to continue long tasks
- **Cline Rules** - Custom .clinerules files for workflows
- **Dual-Mode** - Plan mode (preview) vs Act mode (execute)

### 🔌 Extensibility
- **MCP Integration** - Discover and use MCP servers
- **Custom Tools** - Create domain-specific tools on demand
- **Tool Variants** - Different specs for different LLMs
- **Slack Commands** - Define custom @ mentions

### 🌐 Advanced Capabilities
- **Browser Automation** - Launch browser, click, type, screenshot
- **Terminal Integration** - Execute commands and capture output
- **Web Fetching** - Fetch URLs and convert to markdown
- **Git Integration** - Access git history and diffs

---

## Implementation Patterns

### Tool Handler Pattern (All Tools Follow This)
```
Tool Handler Class
├── name: unique identifier
├── getDescription(): for UI
├── handlePartialBlock(): streaming support
└── execute(): full execution
```

### State Management Pattern
```
StateManager
├── Centralized state storage
├── Atomic updates with persistence
└── Event subscriptions for changes
```

### Context Injection Pattern
```
parseMentions()
├── Regex matching to find mentions
├── Type-specific processing
└── Content transformation to markdown
```

### Approval Pattern
```
AutoApprove.shouldAutoApproveTool()
├── Check YOLO mode
├── Check granular settings
└── Check workspace boundaries
```

---

## Command Reference

### System Commands (Built-in)
- `/newtask` - Create new task with context preload
- `/deep-planning` - Structured investigation and planning
- `/smol` or `/compact` - Condense context window
- `/newrule` - Create custom .clinerules file
- `/reportbug` - Submit bug to GitHub
- `/subagent` - Delegate to CLI subagent

### Proposed Commands for Code Agent
- `/help` - Show available commands
- `/checkpoint save` - Create snapshot
- `/checkpoint list` - List snapshots
- `/checkpoint restore` - Restore snapshot
- `/progress` - Show task progress
- `/models` - List available models
- `/use <model>` - Switch models
- `/add-tool` - Add MCP server
- `/list-tools` - Show available tools
- `/problems` - Show workspace errors
- `/workspace` - Show workspace info
- `/plan` - Switch to plan mode
- `/act` - Switch to act mode
- `/history` - Show task history

---

## Mention System Reference

### Available Mention Types
```
@file or @path/to/file          - Add file content
@folder or @path/to/folder/     - Add all files in folder
@url https://example.com        - Fetch and convert URL
@problems                        - Show workspace diagnostics
@terminal                        - Show latest terminal output
@git-changes or @hash           - Show git diff
@workspace:path/file            - Multi-root workspace support
```

### Example Usage
```
User: "Fix the error in @problems and test with @url https://docs.example.com"

Cline will:
1. Extract all workspace errors
2. Fetch and convert the URL to markdown
3. Include both in context for the agent
```

---

## Auto-Approval Settings Structure

```
autoApprovalSettings: {
  actions: {
    readFiles: boolean                      // @file mentions
    readFilesExternally: boolean            // Outside workspace
    editFiles: boolean                      // Write files
    editFilesExternally: boolean            // Outside workspace
    executeSafeCommands: boolean            // "safe" command list
    executeAllCommands: boolean             // Any command
    useBrowser: boolean                     // Browser automation
    useMcp: boolean                         // MCP tools
  },
  yoloModeToggled: boolean                  // Override all
}
```

---

## Focus Chain (Task Progress) Format

```markdown
# Task Progress

- [x] Initial investigation and planning
- [x] Set up project structure
- [ ] Implement core features
  - [x] Feature A
  - [ ] Feature B
- [ ] Testing and validation
- [ ] Documentation

## Current Status
Working on Feature B implementation

## Next Steps
1. Complete Feature B
2. Run test suite
3. Update documentation
```

---

## Checkpoint Workflow

```
User starts task
    ↓
Agent creates checkpoint #1
    ↓
Agent works and makes changes
    ↓
User says "save checkpoint"
    ↓
Agent creates checkpoint #2
    ↓
Agent continues and breaks something
    ↓
User says "restore checkpoint #2"
    ↓
Workspace reverted to checkpoint #2
    ↓
Agent tries different approach
```

---

## Deep Planning Workflow

```
User: "Implement authentication system"
    ↓
Agent: [Silent Investigation]
  - Analyzes existing code
  - Finds similar patterns
  - Checks dependencies
    ↓
Agent: [Asks Questions]
  "Do you want JWT or session-based?"
  "Any existing auth libraries?"
    ↓
Agent: [Generates Plan]
  Creates implementation_plan.md
  - Overview
  - Type definitions
  - Files to modify
  - Implementation order
    ↓
User: Reviews plan, says "proceed"
    ↓
Agent: Creates task with progress list
    ↓
Agent: Implements step-by-step
```

---

## Tool Registration Pattern

```
Tool Definition (spec.ts)
  ↓
Tool Handler (handler.ts)
  ↓
Register in ToolExecutorCoordinator
  ↓
Add to enhanced_prompt.ts
  ↓
Available to model via system prompt
```

---

## Key Metrics to Track

### For Adoption Priority
1. **Context Savings**: How much does feature reduce token usage?
2. **UX Improvement**: How much does it improve user experience?
3. **Safety Impact**: How much does it improve safety?
4. **Implementation Effort**: How much engineering work is needed?

### Features Ranked by ROI
1. **Mention System** (High UX, Medium effort)
2. **Auto-Approval** (High safety, Low effort)
3. **Task Persistence** (High safety, Low effort)
4. **Checkpoints** (High value, High effort)
5. **Focus Chain** (High context savings, Medium effort)

---

## Integration Checklist

For each feature, ensure:
- [ ] Display layer support (terminal rendering)
- [ ] Session persistence (save/restore)
- [ ] Enhanced prompt awareness (model knows about feature)
- [ ] Error handling (graceful failures)
- [ ] Help/documentation (users understand feature)
- [ ] Tests (unit and integration tests)
- [ ] Settings/config (if applicable)

---

## File Organization

```
features/clines_inspired/
├── draft_log.md                 # Main analysis (you are here)
├── IMPLEMENTATION_EXAMPLES.md   # Code patterns from Cline
├── QUICK_REFERENCE.md           # This file
│
├── features/
│   ├── checkpoints.md           # Checkpoint system design
│   ├── focus_chain.md           # Context compression design
│   ├── mention_system.md        # @mention syntax design
│   ├── auto_approval.md         # Permission system design
│   ├── mcp_integration.md       # MCP tool design
│   └── ...
│
└── implementation/
    ├── checkpoint_implementation.md
    ├── mention_parser.go          (Go example)
    ├── auto_approval.go           (Go example)
    └── ...
```

---

## Research Resources

### Key Cline Folders to Study
- `src/core/task/` - Task execution engine
- `src/core/mentions/` - Mention parsing
- `src/integrations/checkpoints/` - Checkpoint system
- `src/core/prompts/` - Prompt engineering
- `src/core/task/tools/` - Tool handlers
- `src/core/slash-commands/` - Command system

### Key Files to Reference
- `src/core/prompts/commands.ts` - Deep planning logic
- `src/core/task/focus-chain/` - Context compression
- `src/core/task/tools/autoApprove.ts` - Permission system
- `src/core/assistant-message/index.ts` - Tool parsing
- `src/integrations/checkpoints/CheckpointTracker.ts` - State snapshots

---

## Next Steps

1. **Deep Dive**: Pick top 3 features (Checkpoints, Focus Chain, Mentions)
2. **Design Docs**: Create Go-based design for each feature
3. **Prototype**: Implement one feature end-to-end
4. **Get Feedback**: Share prototype with team
5. **Roadmap**: Create implementation timeline
6. **Execute**: Build features in phases

---

## Questions to Answer

- [ ] How does code_agent's display system work?
- [ ] What session persistence mechanism exists?
- [ ] How are tools currently registered/executed?
- [ ] What's the enhanced prompt system?
- [ ] How are approval flows currently handled?
- [ ] Can we leverage existing file watching?
- [ ] What's the terminal capability level?
