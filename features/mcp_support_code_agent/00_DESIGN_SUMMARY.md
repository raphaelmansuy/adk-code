# MCP Support Design Summary

## Executive Summary

This document provides a comprehensive overview of the Model Context Protocol (MCP) support being added to the code_agent CLI tool. The design leverages ADK-Go's production-ready `mcptoolset` abstraction for all MCP protocol handling, paired with custom configuration management and CLI tooling.

**📋 Research Finding**: Research completed Nov 13, 2025 confirms that ADK-Go already provides comprehensive MCP support via `tool/mcptoolset`. See `logs/2025-11-13-09-50_mcp-adk-integration-research.md` for detailed analysis.

## Design Philosophy

**Principle**: Add MCP support as a "tool aggregation" without reimplementing MCP protocol handling.

**Key Decision**: Leverage `google.golang.org/adk/tool/mcptoolset` for MCP client functionality instead of building custom client. This:
- Eliminates ~2000 lines of MCP SDK wrapper code
- Provides production-tested abstractions
- Reduces implementation time from 6-9 weeks to 1-2 weeks
- Maintains consistency with ADK-based architecture

**Approach**: 
1. Use ADK's `mcptoolset` for all MCP protocol handling
2. Build lightweight config and manager wrapper
3. Integrate through existing agent toolset pattern
4. Lazy initialization where possible
5. Clear user-facing diagnostics via CLI commands

## Key Components

### 1. Configuration System (New)
- **Source**: JSON file or environment variable
- **Location**: `~/.code_agent/config.json` (default)
- **Format**: Supports stdio, SSE, HTTP transports
- **Features**: Tool filtering, headers, timeouts
- **Package**: `internal/config/mcp.go`

### 2. MCP Manager (New)
- **Package**: `pkg/mcp/manager.go`
- **Responsibility**: Multi-server orchestration
- **Implementation**: Wraps ADK-Go's `mcptoolset` instances
- **Features**: Server lifecycle, toolset aggregation

### 3. Transport Factory (New)
- **Package**: `pkg/mcp/transport.go`
- **Responsibility**: Create MCP transport objects
- **Transports**: stdio (command), SSE, HTTP
- **Integration**: Passes to `mcptoolset.New()`

### 4. CLI Commands (New)
- **Package**: `internal/cli/commands/mcp.go`
- **Commands**: `/mcp list`, `/mcp tools <server>`, `/mcp reload`
- **Features**: Server status, tool discovery, management

### 5. ADK-Go's mcptoolset (Existing)
- **Package**: `google.golang.org/adk/tool/mcptoolset`
- **Responsibility**: MCP protocol implementation
- **Features**: Tool discovery, session management, execution
- **Why we use it**: Proven, tested, maintained by Google ADK team

## Architectural Benefits

### Over Builtin Tools Only
- **Flexibility**: New tools without code changes
- **Extensibility**: Community-provided MCP servers
- **Scalability**: Unlimited tools via external servers

### Over Custom MCP Client Implementation
- **Maintenance**: MCP SDK updates handled by ADK team
- **Code reduction**: 75% less code to maintain (~800 vs ~5500 lines)
- **Reliability**: Uses proven, tested abstractions
- **Time to market**: 1-2 weeks vs 6-9 weeks

### Over Gemini-CLI TypeScript Implementation
- **Language**: Go-native, no cross-language dependencies
- **Integration**: Seamless with existing ADK-based code_agent
- **Performance**: No JavaScript runtime overhead

### Over ADK-Go's mcptoolset Alone
- **Configuration**: JSON file support with validation
- **Management**: Multi-server coordination and status tracking
- **UX**: CLI commands for user control and debugging

## Implementation Roadmap

### Phase 1: MVP (1-2 weeks)
- Config loading and validation via `internal/config/mcp.go`
- MCP manager with multi-server support via `pkg/mcp/manager.go`
- Transport factory for stdio/SSE/HTTP via `pkg/mcp/transport.go`
- CLI commands (`/mcp list`, `/mcp tools`, `/mcp reload`)
- Comprehensive testing

### Phase 2: Enhanced (1-2 weeks)
- Advanced server lifecycle management
- OAuth2 authentication support
- Tool caching and performance optimization
- Connection health checks

### Phase 3: Future
- Configuration hot-reload
- Resource and Prompt protocol support
- Metrics and monitoring
- YAML config support

## Configuration Strategy

### Simple Path (Single Server)
```bash
# Set environment variable
export CODE_AGENT_MCP_SERVERS='[{
  "name": "filesystem",
  "type": "stdio",
  "command": "mcp-server-filesystem"
}]'

# Run
code-agent
```

### Advanced Path (Multiple Servers)
```bash
# Create config file
cat > ~/.code_agent/config.json << 'EOF'
{
  "mcp": {
    "servers": {
      "filesystem": { ... },
      "github": { ... },
      "web": { ... }
    }
  }
}
EOF

# Run with config
code-agent --config ~/.code_agent/config.json
```

## Security Considerations

### Authentication
- **Headers support**: Custom headers with env var substitution
- **OAuth (Phase 2)**: Automatic OAuth2 flow discovery
- **Token storage**: Secure local storage in `~/.code_agent/mcp-tokens/`

### Trust & Confirmation
- **Tool execution**: Confirmation required by default
- **Trusted servers**: Can skip confirmation if config `trust: true`
- **Tool filtering**: Block dangerous tools via `excludeTools`

## Performance Characteristics

### Startup Time
- **No MCP**: ~100ms (baseline)
- **With 1 server**: ~200-300ms (lazy connect)
- **With 3 servers**: ~500-800ms (parallel discovery)

### Memory Usage
- **Per server**: ~5-10 MB (typical)
- **Tool metadata**: ~100KB per 50 tools
- **Connection pooling**: Reduces memory for multiple tools

### Tool Execution Latency
- **Local (stdio)**: 50-200ms per call
- **Remote (SSE/HTTP)**: 200-2000ms per call (network dependent)

## Error Handling Strategy

### Connection Failures
- **Startup**: Validate config, warn on failures, continue with others
- **Runtime**: Graceful degradation, allow retry
- **Recovery**: `/mcp reconnect <server>` command

### Tool Execution Errors
- **MCP protocol errors**: Surfaced to user
- **Timeout**: Controlled via `timeout` config
- **Partial failures**: Other tools still work

### Configuration Errors
- **Invalid JSON**: Clear error message
- **Missing fields**: Specific suggestion for fix
- **Env vars**: Error if referenced variable not found

## Testing Strategy

### Unit Tests
- Config parsing and validation
- Tool wrapper creation
- Response transformation
- Manager operations

### Integration Tests
- Mock MCP server
- End-to-end tool discovery
- Tool execution workflows
- Error scenarios

### Manual Testing
- Real MCP servers (filesystem, GitHub, etc.)
- Multiple concurrent servers
- Long-running sessions
- Network failures

## Documentation Plan

### User Guides
- `MCP_SETUP.md`: How to configure MCP servers
- `MCP_QUICKSTART.md`: 5-minute quick start
- Configuration examples with real servers

### Developer Guides
- `MCP_DEV.md`: Architecture and design
- `MCP_ARCHITECTURE.md`: Component details
- Extending with custom transports

### Examples
- Example configs for common servers
- Step-by-step setup guides
- Troubleshooting guide

## Success Metrics

### Functionality
- ✅ Configure MCP servers via JSON/environment
- ✅ Discover and register tools from servers
- ✅ Execute MCP tools from agent
- ✅ Show server status and tools
- ✅ Handle errors gracefully

### Code Quality
- ✅ 80%+ test coverage
- ✅ Clear error messages
- ✅ Comprehensive documentation
- ✅ No regressions in existing tools

### User Experience
- ✅ Easy configuration
- ✅ Clear CLI feedback
- ✅ Helpful error messages
- ✅ Good performance

## Comparison with Alternatives

### gemini-cli MCP
| Aspect | gemini-cli | code_agent |
|--------|-----------|-----------|
| Language | TypeScript | Go |
| Config | Complex JSON | Simple JSON |
| OAuth | Built-in | Phase 2 |
| Multi-server | Yes | Yes |
| Tool filter | Yes | Yes |
| CLI commands | Extensive | Essential |
| Code complexity | High | Moderate |

### adk-go MCP
| Aspect | adk-go | code_agent |
|--------|--------|-----------|
| Complexity | Minimal | Simple |
| Config | Programmatic | JSON/env |
| Tool filter | Predicate-based | Config-based |
| Multi-server | Manual | Managed |
| CLI commands | None | Yes |
| OAuth | None | Phase 2 |

### cline MCP
| Aspect | cline | code_agent |
|--------|-------|-----------|
| Language | TypeScript | Go |
| MCP protocol | Full SDK | ADK-based |
| Complexity | High | Moderate |
| Config flexibility | High | Good |
| Error recovery | Good | Good |

## Known Limitations (Phase 1)

1. **No OAuth support** (added in Phase 2)
2. **No tool result caching** (can add later)
3. **No configuration hot-reload** (requires restart)
4. **No resource support** (MCP resource access) (Phase 3)
5. **No prompt support** (MCP prompts) (Phase 3)
6. **Simple response formatting** (basic text/JSON support)

## Future Enhancements

### Near Term (Phase 2)
- OAuth2 auto-discovery and auth flow
- Token refresh and storage
- Server health checks
- Connection retries with backoff

### Medium Term (Phase 3)
- YAML config file support
- Configuration hot-reload
- Tool result caching
- Advanced filtering UI

### Long Term
- Resource protocol support
- Prompt protocol support
- Metrics and monitoring
- Performance optimizations

## Migration Path

### From Existing Setups
1. **Cline users**: Export MCP servers from Cline config
2. **Standalone tools**: Register as stdio MCP servers
3. **Cloud APIs**: Wrap with MCP servers, use via config

### From Other Tools
1. Extract MCP server definitions
2. Convert to code_agent format
3. Test with `--validate-config`
4. Deploy via JSON or environment

## Getting Started

### For Users
1. Read `docs/MCP_SETUP.md`
2. Create `~/.code_agent/config.json`
3. Start `code-agent --config ~/.code_agent/config.json`
4. Use `/mcp list-tools` to see available tools

### For Developers
1. Read `docs/MCP_DEV.md` for architecture
2. Review `01_MCP_SPECIFICATION.md` for design
3. Follow `02_IMPLEMENTATION_PLAN.md` for phased implementation
4. Reference `03_CONFIGURATION_FORMAT.md` for config details

## Support & Feedback

### Reporting Issues
1. Check troubleshooting guide
2. Run with debug: `CODE_AGENT_MCP_DEBUG=1`
3. Include config (sanitized) and error logs
4. Test with mock server if possible

### Contributing
1. Feature requests welcome
2. New MCP servers as examples
3. Config format feedback
4. Performance improvements

## Conclusion

The MCP support design balances:
- **Simplicity**: Phase 1 covers 80% of use cases
- **Extensibility**: Easy to add features as needed
- **Robustness**: Comprehensive error handling
- **Integration**: Works seamlessly with code_agent

The phased approach allows:
- Quick MVP delivery (Phase 1)
- Production-quality features (Phase 2)
- Advanced capabilities (Phase 3)

Implementation can begin immediately with clear specification, test infrastructure, and documentation ready to support development.

---

## Document Reference

| Document | Purpose |
|----------|---------|
| `01_MCP_SPECIFICATION.md` | Technical specification of MCP support |
| `02_IMPLEMENTATION_PLAN.md` | Detailed sprint-by-sprint implementation roadmap |
| `03_CONFIGURATION_FORMAT.md` | Configuration file format guide with examples |
| `draft_notepad_log.md` | Raw research notes from gemini-cli and adk-go |
| `thought_notepad_log.md` | Design thinking and decision rationale |
| This document | Executive summary and overview |
