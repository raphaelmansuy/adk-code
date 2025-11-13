# README: Code Agent Documentation Suite

Complete documentation for the Code Agent codebase. Start here.

---

## 📚 Documentation Overview

This documentation suite provides **multiple entry points** for understanding Code Agent, from high-level architecture to implementation details.

### Which Document Should I Read?

```
Are you new to Code Agent?
  ↓
  Start with QUICK_REFERENCE.md (2 min read)
  
Want to understand the system design?
  ↓
  Read ARCHITECTURE.md (15 min read)
  
Want to implement a new feature?
  ↓
  Read TOOL_DEVELOPMENT.md (20 min + implementation)
  
Need deep technical analysis?
  ↓
  Read draft.md (reference document)
```

---

## 📖 Document Guide

### 1. QUICK_REFERENCE.md
**Best for**: Daily use, common commands, quick lookups  
**Time**: 2-3 minutes  
**Contains**:
- Building & running the app
- Environment setup (Gemini, OpenAI, Vertex AI)
- CLI flags reference
- In-REPL commands
- Project structure
- Common issues & solutions
- Key file locations

**When to use**: You're ready to **start using** Code Agent

---

### 2. ARCHITECTURE.md
**Best for**: Understanding system design and component interaction  
**Time**: 15-20 minutes  
**Contains**:
- System architecture overview (with diagrams)
- 4-part component system (Display, Model, Agent, Session)
- Detailed component analysis
- Application lifecycle
- Tool ecosystem overview
- Configuration & environment
- Error handling strategy
- Design patterns (Builder, Adapter, Composition)
- Testing strategy
- Performance considerations
- Extensibility guide

**When to use**: You want to **understand the big picture** or **extend the system**

---

### 3. TOOL_DEVELOPMENT.md
**Best for**: Creating new tools for the agent  
**Time**: 20 minutes + implementation  
**Contains**:
- Complete 4-step tool pattern
- Step-by-step example (CountLines tool)
- Tool pattern templates (file ops, execution, analysis)
- Safety considerations & input validation
- Testing templates
- Common patterns & recipes
- Troubleshooting guide
- Integration with agent

**When to use**: You're ready to **implement a new tool**

---

### 4. draft.md
**Best for**: Deep technical reference  
**Time**: 30-40 minutes (reference document)  
**Contains**:
- Project overview & value proposition
- Architecture patterns (Builder, Component Composition, Orchestrator)
- Tool ecosystem analysis
- Model & LLM abstraction
- Comprehensive internal package map
- Key design decisions
- Data flows (user interaction, model selection, token tracking)
- Key files to understand
- Conventions & patterns
- External dependencies
- Strengths & observations
- Summary statistics

**When to use**: You need **deep technical reference** or want to **contribute to core systems**

---

## 🎯 Learning Paths

### Path 1: User (5 Minutes)
1. Read QUICK_REFERENCE.md sections:
   - Building & Running
   - In-REPL Commands
2. Run `make build && make run`
3. Try some commands

**Result**: Can use Code Agent effectively

---

### Path 2: Contributor (1 Hour)
1. Read QUICK_REFERENCE.md (5 min)
2. Read ARCHITECTURE.md sections:
   - System Architecture Overview (5 min)
   - Component Architecture (10 min)
3. Run and explore code (10 min)
4. Read TOOL_DEVELOPMENT.md (20 min)

**Result**: Can implement new tools

---

### Path 3: Core Contributor (3 Hours)
1. Complete Path 2 (1 hour)
2. Read draft.md (30 min)
3. Read ARCHITECTURE.md sections:
   - Tool Ecosystem (10 min)
   - Design Patterns (15 min)
   - Extensibility (10 min)
4. Explore codebase (30 min)
5. Study key files in recommended order (30 min)

**Result**: Can modify core systems (orchestration, display, session management)

---

## 🗺️ Codebase Map

### Critical Files (Read in Order)

```
1. main.go (140 lines)
   ↓
2. internal/orchestration/builder.go (140 lines)
   ↓
3. internal/app/app.go (140 lines)
   ↓
4. internal/repl/repl.go (245 lines)
   ↓
5. tools/file/read_tool.go (130 lines) [tool pattern]
   ↓
6. pkg/models/registry.go (218 lines) [model selection]
   ↓
7. internal/display/renderer.go [UI rendering]
```

**Total critical code**: ~1000 lines (highly scalable for learning)

### Package Organization

```
adk-code/
├── main.go                          Entry point
├── internal/
│   ├── app/                         Application lifecycle
│   ├── orchestration/               Component builder pattern
│   ├── repl/                        Interactive loop
│   ├── display/                     Terminal UI (8 subpackages)
│   │   ├── renderer/
│   │   ├── streaming/
│   │   ├── banner/
│   │   ├── components/
│   │   ├── formatters/
│   │   ├── styles/
│   │   ├── terminal/
│   │   └── tools/
│   ├── session/                     Persistence + token tracking
│   ├── config/                      Configuration loading
│   ├── cli/                         Built-in commands
│   ├── llm/                         LLM provider abstraction
│   ├── runtime/                     Signal handling
│   ├── tracking/                    Token tracking
│   └── prompts/                     System prompts
├── pkg/
│   ├── models/                      Model registry + factories
│   ├── errors/                      Error handling
│   ├── workspace/                   Path resolution
│   └── testutil/                    Test utilities
└── tools/ (8 categories)
    ├── file/                        File operations
    ├── edit/                        Code editing
    ├── exec/                        Command execution
    ├── search/                      Search & discovery
    ├── display/                     Agent→UI messaging
    ├── workspace/                   Workspace analysis
    ├── v4a/                         V4A patch format
    └── base/                        Registry & errors
```

---

## 🔧 Development Workflow

### Setting Up Local Development

```bash
# 1. Clone and navigate
cd /path/to/adk_training_go

# 2. Build
cd code_agent
make build

# 3. Run tests
make test

# 4. Run quality checks (before committing)
make check
```

### Creating a New Tool

```bash
# 1. Read TOOL_DEVELOPMENT.md
# 2. Create tools/CATEGORY/new_tool.go
# 3. Implement 4-step pattern
# 4. Test
cd code_agent
make test
# 5. Run full quality check
make check
```

### Understanding a Component

```bash
# 1. Start with key file list (above)
# 2. Read ARCHITECTURE.md relevant section
# 3. Study the actual code in adk-code/
# 4. Run the app and test the component
# 5. Check tests in *_test.go files
```

---

## 📊 Project Statistics

| Metric | Value |
|--------|-------|
| Total Go packages | 20+ |
| Internal packages | 11 |
| Tool categories | 8 |
| Total tools | ~30 |
| Critical code lines | ~1000 |
| Supported LLM backends | 3 (Gemini, Vertex AI, OpenAI) |
| Built-in REPL commands | 6+ |
| Max context window | 1,000,000 tokens (Gemini 2.5 Flash) |

---

## 🎓 Key Concepts

### Architecture Patterns

| Pattern | Used For | Location |
|---------|----------|----------|
| **Builder (Orchestrator)** | Component wiring | `internal/orchestration/` |
| **Tool Factory** | Tool creation | `tools/*/xxx_tool.go` |
| **Adapter** | LLM backend abstraction | `pkg/models/` |
| **Component Composition** | System modularity | Application struct |
| **Strategy** | Output formatting | `internal/display/formatters/` |

### Design Principles

1. **Composition over Inheritance**: 4 independent components
2. **Single Responsibility**: Each package has clear purpose
3. **Dependency Injection**: Config passed to components
4. **Type Safety**: JSON schema validation for tools
5. **Error Handling**: Output structs, never panics
6. **Testability**: Mocks available for each component

---

## 🚀 Common Tasks

### "I want to use Code Agent"
→ Read QUICK_REFERENCE.md → Run `make build && make run`

### "I want to add a new tool"
→ Read TOOL_DEVELOPMENT.md → Follow 4-step pattern

### "I want to support a new LLM backend"
→ Read ARCHITECTURE.md (Model Subsystem) → Implement adapter

### "I want to add a new output format"
→ Read ARCHITECTURE.md (Display Subsystem) → Create formatter

### "I want to understand the code"
→ Follow Learning Path 3 above

### "I want to debug an issue"
→ See QUICK_REFERENCE.md Debugging Tips section

---

## 📝 Documentation Maintenance

These documents are **maintained** and **kept in sync** with the codebase.

When modifying Code Agent:
1. Update relevant code
2. Update corresponding documentation
3. Verify examples still work
4. Run `make check` to ensure tests pass

---

## 🔗 Navigation Quick Links

- **Getting Started**: QUICK_REFERENCE.md → "Building & Running"
- **System Design**: ARCHITECTURE.md → "System Architecture Overview"
- **Creating Tools**: TOOL_DEVELOPMENT.md → "Step-by-Step Example"
- **Deep Dive**: draft.md → "Architecture Patterns"
- **Troubleshooting**: QUICK_REFERENCE.md → "Common Issues & Solutions"
- **Learning**: This file → "Learning Paths"

---

## 💡 Pro Tips

1. **Start small**: Read QUICK_REFERENCE.md first, not the whole suite
2. **Hands-on**: Run code while reading (e.g., `make run` while reading ARCHITECTURE.md)
3. **Example-driven**: TOOL_DEVELOPMENT.md has a complete working example
4. **Reference mode**: Keep draft.md open while exploring the codebase
5. **Incremental**: Each document builds on previous ones

---

## 🤔 FAQ

**Q: Where should I start?**  
A: Start with QUICK_REFERENCE.md, then run the app.

**Q: How do I understand the system?**  
A: Read ARCHITECTURE.md (15 minutes), then explore the code.

**Q: How do I add a tool?**  
A: Read TOOL_DEVELOPMENT.md and follow the 4-step example.

**Q: What's the most important file?**  
A: main.go (140 lines) - everything else follows from there.

**Q: How long will it take to learn?**  
A: 1 hour for basics, 3 hours for core contribution readiness.

**Q: Where's the API documentation?**  
A: Not in these docs - instead, see function signatures in code with inline comments.

**Q: Can I modify the codebase?**  
A: Yes! Follow the 4-step tool pattern (TOOL_DEVELOPMENT.md) or study existing code before making changes.

---

## 📞 Support

For detailed information, refer to:
- **Code Questions**: ARCHITECTURE.md or draft.md
- **Tool Development**: TOOL_DEVELOPMENT.md
- **Getting Started**: QUICK_REFERENCE.md
- **Running Code**: Makefile targets + QUICK_REFERENCE.md

---

## 📄 Document Status

**Last Updated**: November 12, 2025  
**Status**: Complete and comprehensive  
**Coverage**: Architecture, tool development, quick reference, deep analysis  
**Accuracy**: Synchronized with codebase as of this date

---

## 🎓 Next Steps After Reading

1. ✅ Read QUICK_REFERENCE.md (2 min)
2. ✅ Build the app: `make build`
3. ✅ Run the app: `../bin/adk-code`
4. ✅ Try a command: `How do I read a file?`
5. ✅ Read ARCHITECTURE.md (15 min)
6. ✅ Explore a tool: `adk-code/tools/file/read_tool.go`
7. ✅ Read TOOL_DEVELOPMENT.md (20 min)
8. ✅ Create your first tool (30 min)

**After these steps, you'll have:**
- Working understanding of the system
- Ability to use Code Agent effectively
- Skills to implement new tools
- Knowledge to extend the codebase

Happy learning! 🚀

