# Session Persistence Feature - Quick Start Guide

## Overview
The code_agent now supports persistent sessions using SQLite. This means:
- ✅ Conversation history persists across application restarts
- ✅ Multiple parallel sessions can be maintained
- ✅ Easy session management via CLI commands

## Installation/Update
No special installation needed - just rebuild:
```bash
make build
```

## Basic Usage

### Create a new session
```bash
./code-agent new-session my-project
✨ Created new session: my-project
```

### List all sessions
```bash
./code-agent list-sessions
📋 Sessions:
1. my-project (5 events)
2. testing (2 events)
```

### Resume a session
```bash
./code-agent --session my-project
📖 Resumed session: my-project (5 events)
```

### Delete a session
```bash
./code-agent delete-session my-project
🗑️  Deleted session: my-project
```

### Default session
If no session is specified, uses "default":
```bash
./code-agent  # Uses session named "default"
```

### Custom database location
```bash
./code-agent --db /path/to/sessions.db --session my-session
```

## Session Storage
- Default location: `~/.code_agent/sessions.db`
- SQLite database with persistent storage
- Automatic schema creation on first run

## Architecture
New `code_agent/persistence/` package provides:
- **SQLiteSessionService**: Implements ADK's session.Service interface
- **SessionManager**: High-level session management API
- **Models**: Data models for sessions, events, and state

## Files Changed
- `main.go`: Added CLI commands and session integration
- `go.mod`: Added GORM and SQLite driver dependencies
- `persistence/`: New package for session storage

## Under the Hood
- Sessions use SQLite with GORM ORM
- Full ACID transactions for consistency
- Support for app-level, user-level, and session-level state
- Event history stored with full LLM response data

## API for Developers
```go
// Create session manager
manager, err := persistence.NewSessionManager("app-name", "/path/to/db")
defer manager.Close()

// Create session
session, err := manager.CreateSession(ctx, userID, sessionName)

// List sessions
sessions, err := manager.ListSessions(ctx, userID)

// Get specific session
session, err := manager.GetSession(ctx, userID, sessionName)

// Delete session
err := manager.DeleteSession(ctx, userID, sessionName)
```

## Backwards Compatibility
Existing workflows work without changes:
```bash
./code-agent  # Still works, uses "default" session
```

## Troubleshooting

### Database locked
This typically means multiple processes are accessing the same database. Ensure only one instance of code-agent is running per database file.

### Session not found
Session names are case-sensitive. Verify the name with:
```bash
./code-agent list-sessions
```

### Clear all sessions
```bash
rm ~/.code_agent/sessions.db
```

## Next Steps
- Sessions can be extended with branching, snapshots, or export features
- Consider adding session search/filtering capabilities
- Potential for cloud sync or backup functionality

---

**For detailed implementation details, see:** `logs/2025-11-10-session-persistence-implementation.md`
