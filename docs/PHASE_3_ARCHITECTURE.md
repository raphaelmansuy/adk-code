# Phase 3 Architecture Design

**Date:** November 14, 2025  
**Version:** 1.0 (Draft)  
**Status:** 📋 Planning  
**Last Updated:** November 14, 2025  

---

## 1. System Overview

### 1.1 Phase 3 System Architecture

Phase 3 extends the Phase 2 architecture with three major subsystems:

```
┌──────────────────────────────────────────────────────────────────┐
│                        User Interface                             │
│                  (REPL + Headless Mode)                          │
└──────────────────┬───────────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ↓                     ↓
   Interactive Mode      Headless Mode (NEW)
        │                     │
        └──────────┬──────────┘
                   │
┌──────────────────┴───────────────────────────────────────────────┐
│              Orchestrator (Phase 2 + Extensions)                 │
│  • Agent lifecycle management                                    │
│  • Tool invocation coordination                                  │
│  • Session and context management                               │
└──────────────────┬───────────────────────────────────────────────┘
                   │
    ┌──────────────┼──────────────┬──────────────┐
    │              │              │              │
    ↓              ↓              ↓              ↓
┌─────────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐
│   LLM   │  │ Tool Reg │  │ Session    │  │ Metrics  │
│ Backend │  │ (Phase2) │  │ Mgmt       │  │ (NEW)    │
└────┬────┘  └──────┬───┘  └──────┬─────┘  └──────────┘
     │              │             │
     │     ┌────────┴─────────────┤
     │     │                      │
     │     ↓                      ↓
     │  ┌─────────────────────────────────────┐
     │  │  Tool Execution Layer (NEW)         │
     │  │  ┌─────────────────────────────────┐│
     │  │  │ Standard Tools (Phase 2)        ││
     │  │  └────┬────────────────────────────┘│
     │  │       │                              │
     │  │  ┌────┴───────────────────────────┐ │
     │  │  │ Execution Strategies (NEW)     │ │
     │  │  │ • Direct Execution             │ │
     │  │  │ • Docker Sandboxing (NEW)      │ │
     │  │  │ • Remote/SSH (future)          │ │
     │  │  └────────────────────────────────┘ │
     │  └─────────────────────────────────────┘
     │
     ↓
┌──────────────────────────────────────────────────────────────────┐
│          Extensibility Layer (NEW)                               │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐  │
│  │ MCP Client     │  │ Plugin System  │  │ Custom Tools     │  │
│  │ (NEW)          │  │ (NEW)          │  │ (NEW)            │  │
│  └────────────────┘  └────────────────┘  └──────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
     │
     ↓
┌──────────────────────────────────────────────────────────────────┐
│          Infrastructure & Utilities (NEW)                        │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐  │
│  │ Credentials    │  │ Audit Logging  │  │ Context Compress │  │
│  │ Management     │  │                │  │                  │  │
│  └────────────────┘  └────────────────┘  └──────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 Key Design Principles

1. **Layered Architecture**: Clear separation between execution strategies, tool management, and extensibility
2. **Backward Compatibility**: Phase 3 builds on Phase 2 without breaking existing code
3. **Pluggable Execution**: Multiple execution strategies (direct, docker, future: SSH/remote)
4. **Isolation by Default**: Security-first design with sandboxing capabilities
5. **Observable & Auditable**: Complete logging and metrics for compliance
6. **Extensible by Design**: Plugins and MCP as first-class citizens

---

## 2. Component Specifications

### 2.1 Docker Sandboxing (Phase 3.1)

#### Architecture

```go
┌────────────────────────────────────────┐
│  Tool Invocation (from ADK Framework)  │
└─────────────────┬──────────────────────┘
                  │
                  ↓
         ┌────────────────┐
         │  ExecutionMux  │  (Router)
         │ (StrategyPick) │
         └────────┬───────┘
                  │
        ┌─────────┴─────────┐
        │                   │
        ↓                   ↓
   DirectExecutor    DockerExecutor (NEW)
        │                   │
        ├─→ OS process      ├─→ Docker Client
        │                   │
        │                   ├─→ Image Pull/Run
        │                   │
        │                   ├─→ Container Mgmt
        │                   │
        │                   └─→ Resource Limits
        │
        └─────────┬─────────┘
                  │
                  ↓
         ┌────────────────┐
         │  ExecutionResult│  (Unified)
         └────────────────┘
```

#### Component: DockerExecutor

```go
type DockerExecutor struct {
    // Client and config
    Client      *docker.Client
    Config      DockerConfig
    
    // Lifecycle
    Container   *docker.Container
    ImageName   string
    
    // Execution parameters
    Command     []string
    EnvVars     map[string]string
    WorkDir     string
    
    // Monitoring
    Logger      AuditLogger
    Metrics     ExecutionMetrics
}

type DockerConfig struct {
    ImageName       string              // e.g., "golang:1.24"
    ImagePullPolicy ImagePullPolicy     // Always/IfNotPresent/Never
    
    ResourceLimits  ResourceLimits      // CPU, Memory, Disk
    NetworkMode     string              // "none", "bridge", "host"
    VolumeMounts    []VolumeMount
    
    TimeoutSeconds  int
    RetryPolicy     RetryPolicy
}

type ResourceLimits struct {
    CPUShares       int64               // 0 = unlimited
    MemoryMB        int64               // Memory limit
    MemorySwapMB    int64               // Swap limit
    PidsLimit       int64               // Process limit
    DiskQuotaGB     int64               // Volume quota
}

type VolumeMount struct {
    HostPath        string              // /Users/me/workspace
    ContainerPath   string              // /workspace
    ReadOnly        bool
}
```

#### Execution Flow

```
1. PreExecution
   ├─ Validate container image
   ├─ Pull image if needed
   ├─ Prepare volumes
   └─ Setup credentials (masked env vars)

2. CreateContainer
   ├─ Create config with resource limits
   ├─ Mount volumes
   ├─ Inject environment
   └─ Return container ID

3. RunContainer
   ├─ Start container
   ├─ Stream logs to audit
   ├─ Monitor resource usage
   ├─ Handle signals/timeout
   └─ Capture exit code

4. Cleanup
   ├─ Stop container
   ├─ Remove container
   ├─ Cleanup volumes (optional)
   └─ Log completion

5. PostExecution
   ├─ Parse output
   ├─ Record metrics
   ├─ Return ExecutionResult
   └─ Archive audit log
```

#### Error Handling

```go
type DockerError struct {
    Type    DockerErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

const (
    ErrorImageNotFound      DockerErrorType = iota
    ErrorContainerFailed
    ErrorTimeout
    ErrorResourceExhausted
    ErrorNetworkError
    ErrorInvalidConfig
)
```

### 2.2 Credential Management (Phase 3.1)

#### Architecture

```
API Keys / Secrets (from config)
         │
         ↓
┌────────────────────────────┐
│ CredentialManager          │
│ • Store secrets            │
│ • Mask in logs             │
│ • Inject into context      │
└────────────────────────────┘
         │
    ┌────┴─────┐
    │           │
    ↓           ↓
 Direct      Docker (via env vars with masking)
 Injection
```

#### Component: CredentialManager

```go
type CredentialManager struct {
    // Storage
    Credentials map[string]Secret
    
    // Vault integration (optional)
    VaultConfig *VaultConfig
    VaultClient *VaultClient
    
    // Masking
    MaskPatterns []string  // Patterns to mask in logs
    MaskValue    string    // "[REDACTED]"
}

type Secret struct {
    Name        string          // "GITHUB_TOKEN"
    Value       string          // Token value
    Type        SecretType      // API_KEY, CREDENTIAL, etc.
    CreatedAt   time.Time
    ExpiresAt   time.Time
    
    // Metadata
    Source      string          // "environment", "vault", "config"
    Scopes      []string        // ["github", "git"]
}

type VaultConfig struct {
    Address         string              // "https://vault.example.com"
    Token           string              // Auth token
    EngineType      string              // "kv", "generic"
    MountPath       string              // "/secret"
}
```

#### Methods

```go
// Store a credential
func (cm *CredentialManager) StoreSecret(secret Secret) error

// Retrieve a credential
func (cm *CredentialManager) GetSecret(name string) (*Secret, error)

// Inject secrets into environment for execution
func (cm *CredentialManager) InjectIntoEnv(envVars map[string]string) map[string]string

// Mask secrets in output/logs
func (cm *CredentialManager) Mask(text string) string

// Validate secret access permissions
func (cm *CredentialManager) ValidateAccess(secret string, scopes []string) error
```

### 2.3 Audit Logging (Phase 3.1)

#### Architecture

```
All Execution Activity
    │
    ├─ Commands executed
    ├─ Output captured
    ├─ Errors logged
    ├─ Duration tracked
    └─ Resources used
    │
    ↓
┌────────────────────────┐
│ AuditLogger            │
│ • Structure events     │
│ • Mask secrets         │
│ • Write to storage     │
└────────────────────────┘
    │
    ├─ File (JSON)
    ├─ Stdout
    ├─ Webhook
    └─ Future: Remote syslog
```

#### Component: AuditLogger

```go
type AuditLogger struct {
    // Configuration
    Config      AuditConfig
    FilePath    string
    
    // State
    Events      []AuditEvent
    Buffer      *bytes.Buffer
    
    // Monitoring
    EventCount  int
    LastFlush   time.Time
}

type AuditEvent struct {
    // Core fields
    ID          string          // UUID
    Timestamp   time.Time
    EventType   AuditEventType
    
    // Execution context
    ExecutionID string
    AgentName   string
    
    // Event details
    Details     map[string]interface{}
    
    // Security
    Masked      bool
    Sensitive   []string  // Fields that were masked
}

type AuditEventType string

const (
    EventTypeExecutionStart     AuditEventType = "execution_start"
    EventTypeCommandExecuted                   = "command_executed"
    EventTypeOutputCaptured                    = "output_captured"
    EventTypeErrorOccurred                     = "error_occurred"
    EventTypeExecutionEnd                      = "execution_end"
    EventTypeCredentialAccess                  = "credential_access"
)

type AuditConfig struct {
    // Output
    FilePath        string          // "/var/log/adk-code/audit.json"
    MaxFileSizeMB   int
    MaxBackups      int
    MaxAgeDays      int
    
    // Filtering
    LogLevel        string          // "info", "debug", "trace"
    IncludeOutput   bool            // Include stdout in audit
    IncludeErrors   bool            // Include stderr in audit
    
    // Remote
    WebhookURL      string          // Optional remote endpoint
    WebhookHeaders  map[string]string
}
```

#### Methods

```go
func (al *AuditLogger) LogExecutionStart(execID string, agent *Agent) error
func (al *AuditLogger) LogCommand(execID string, cmd string, args []string) error
func (al *AuditLogger) LogOutput(execID string, output string) error
func (al *AuditLogger) LogError(execID string, err error) error
func (al *AuditLogger) LogExecutionEnd(execID string, exitCode int, duration time.Duration) error
func (al *AuditLogger) LogCredentialAccess(execID string, credName string, scopes []string) error
func (al *AuditLogger) Flush() error
func (al *AuditLogger) Query(filter AuditFilter) ([]AuditEvent, error)
```

---

### 2.4 Headless Mode & Batch Execution (Phase 3.2)

#### Architecture

```
┌─────────────────────────────────────┐
│    Headless Mode (NEW)              │
│  Non-interactive, fully automated   │
└──────────────┬──────────────────────┘
               │
      ┌────────┴────────┐
      │                 │
      ↓                 ↓
Batch Input       Configuration
(JSON/YAML)       (BatchConfig)
      │                 │
      └────────┬────────┘
               │
               ↓
     ┌─────────────────────┐
     │ BatchExecutor       │
     │ • Parse input       │
     │ • Validate params   │
     │ • Execute tasks     │
     │ • Format output     │
     └──────────┬──────────┘
                │
      ┌─────────┴─────────┐
      │                   │
      ↓                   ↓
  JSON Output         CSV/Text
  (structured)        (formatted)
```

#### Component: BatchExecutor

```go
type BatchExecutor struct {
    Config      BatchConfig
    Orchestrator *Orchestrator
    Logger      *log.Logger
}

type BatchConfig struct {
    // Execution
    TimeoutSeconds  int
    MaxRetries      int
    ContinueOnError bool
    
    // Input format
    InputFormat     string  // "json", "yaml"
    
    // Output
    OutputFormat    string  // "json", "text", "csv"
    OutputFile      string  // Empty = stdout
    PrettyPrint     bool
    
    // Logging
    LogFile         string
    LogLevel        string  // "debug", "info", "warn", "error"
}

type BatchInput struct {
    // Core
    Query       string
    AgentName   string  // Optional, uses default if empty
    
    // Parameters
    Parameters  map[string]interface{}
    
    // Options
    Timeout     int     // Seconds, overrides config
    OutputFile  string  // Overrides config
    
    // Metadata
    Tags        []string
    Metadata    map[string]interface{}
}

type BatchOutput struct {
    // Result
    Success     bool
    Result      string
    
    // Status
    Status      string      // "success", "error", "timeout"
    ExitCode    int
    
    // Timing
    StartTime   time.Time
    EndTime     time.Time
    Duration    time.Duration
    
    // Metadata
    ExecutionID string
    AgentName   string
    Tags        []string
    Metadata    map[string]interface{}
}
```

#### Methods

```go
func (be *BatchExecutor) ExecuteBatch(ctx context.Context, inputs []BatchInput) ([]BatchOutput, error)
func (be *BatchExecutor) ExecuteOne(ctx context.Context, input BatchInput) (*BatchOutput, error)
func (be *BatchExecutor) ValidateInput(input BatchInput) error
func (be *BatchExecutor) FormatOutput(outputs []BatchOutput) (string, error)
```

### 2.5 Session Management & Checkpointing (Phase 3.2)

#### Architecture

```
Session Events
    │
    ├─ Message added
    ├─ Tool called
    ├─ Tokens counted
    └─ Context updated
    │
    ↓
┌──────────────────────────────────────┐
│ Session Checkpoint System            │
│ • Monitor token usage                │
│ • Detect 75% context limit           │
│ • Trigger compression                │
└──────────────────────────────────────┘
    │
    ├─ Summarize old messages
    ├─ Compress context
    ├─ Save checkpoint
    └─ Continue execution
```

#### Component: SessionCheckpoint

```go
type SessionCheckpoint struct {
    // Metadata
    SessionID       string
    CheckpointID    string      // UUID
    CreatedAt       time.Time
    
    // Context state
    TotalTokens     int
    UsedTokens      int
    TokenThreshold  float64     // 0.75 = 75%
    
    // Compression
    MessageCount    int
    MessagesSummary string      // Compressed summary
    CompressedAt    time.Time
    
    // Recovery
    ContextData     []byte      // Serialized context
    SessionState    map[string]interface{}
}

type ContextCompressor struct {
    // Thresholds
    CompressionThreshold    float64         // 0.75
    SummarizationThreshold  float64         // 0.85
    
    // Strategy
    MaxMessagesToKeep       int             // Keep recent N messages
    SummaryLength           int             // Max summary chars
    
    // State
    Logger                  *log.Logger
}
```

#### Compression Strategy

```go
// When usage > 75% threshold:
// 1. Identify old messages (before T-30 minutes)
// 2. Summarize them into a single system message
// 3. Remove old messages from context
// 4. Save checkpoint with compression data
// 5. Continue execution with reduced tokens

func (cc *ContextCompressor) ShouldCompress(session *Session) bool {
    used := session.GetTokenUsage()
    limit := session.GetContextLimit()
    return float64(used) / float64(limit) > cc.CompressionThreshold
}

func (cc *ContextCompressor) CompressContext(session *Session) (*ContextCheckpoint, error) {
    // Get old messages
    messages := session.GetMessagesBeforeTime(time.Now().Add(-30 * time.Minute))
    
    // Summarize
    summary, err := cc.summarizeMessages(messages)
    if err != nil {
        return nil, err
    }
    
    // Remove old messages
    for _, msg := range messages {
        session.RemoveMessage(msg.ID)
    }
    
    // Insert summary as system message
    session.AddMessage(&Message{
        Role: "system",
        Content: "Previous conversation summary: " + summary,
    })
    
    // Create checkpoint
    checkpoint := &ContextCheckpoint{
        SessionID: session.ID,
        CompressedMessages: len(messages),
        Summary: summary,
    }
    
    return checkpoint, nil
}
```

### 2.6 MCP Support (Phase 3.3)

#### Architecture

```
┌──────────────────────────┐
│  MCP Server              │
│  (External Process)      │
│  • Defines tools         │
│  • Listens on socket     │
└────────────┬─────────────┘
             │
             │ (stdio/socket)
             │
             ↓
┌──────────────────────────┐
│  MCP Client              │
│  • Connect to server     │
│  • Discover tools        │
│  • Call tools            │
│  • Handle errors         │
└────────────┬─────────────┘
             │
             ↓
┌──────────────────────────┐
│  Tool Registry           │
│  • Register MCP tools    │
│  • Route invocations     │
│  • Cache tool metadata   │
└────────────┬─────────────┘
             │
             ↓
┌──────────────────────────┐
│  ADK Framework           │
│  • Use as standard tools │
│  • Full integration      │
└──────────────────────────┘
```

#### Component: MCPClient

```go
type MCPClient struct {
    // Connection
    ServerAddress   string
    ProcessType     string  // "stdio", "sse", "http"
    Connection      interface{}
    
    // Tools cache
    Tools           map[string]*MCPTool
    ToolsLastSync   time.Time
    
    // Lifecycle
    IsConnected     bool
    RetryPolicy     RetryPolicy
    
    // Logging
    Logger          *log.Logger
}

type MCPTool struct {
    // Identity
    Name            string
    Description     string
    
    // Schemas
    InputSchema     *JSONSchema
    OutputSchema    *JSONSchema
    
    // Metadata
    ServerName      string
    ServerVersion   string
}

type JSONSchema struct {
    Type        string                          // "object", "string", etc.
    Properties  map[string]*JSONSchema
    Required    []string
    Description string
}
```

#### Methods

```go
func (c *MCPClient) Connect(ctx context.Context) error
func (c *MCPClient) Disconnect() error
func (c *MCPClient) DiscoverTools() ([]MCPTool, error)
func (c *MCPClient) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)
func (c *MCPClient) GetToolMetadata(toolName string) (*MCPTool, error)
func (c *MCPClient) Health() error
```

### 2.7 Plugin Architecture (Phase 3.3)

#### Architecture

```
Plugin Directory
    │
    ├─ plugin-a/plugin.yaml
    ├─ plugin-b/plugin.yaml
    └─ plugin-c/plugin.yaml
    │
    ↓
┌────────────────────────────┐
│ Plugin Loader              │
│ • Discover plugins         │
│ • Validate manifests       │
│ • Check dependencies       │
│ • Load code                │
└────────────┬───────────────┘
             │
             ↓
┌────────────────────────────┐
│ Plugin Registry            │
│ • Track plugins            │
│ • Manage lifecycle         │
│ • Route invocations        │
│ • Handle unloading         │
└────────────┬───────────────┘
             │
             ↓
┌────────────────────────────┐
│ Custom Tool Factory        │
│ • Wrap plugin functions    │
│ • Validate inputs          │
│ • Stream outputs           │
└────────────┬───────────────┘
             │
             ↓
┌────────────────────────────┐
│ Tool Registry (Phase 2)    │
│ • Unified tool access      │
│ • ADK integration          │
└────────────────────────────┘
```

#### Component: Plugin

```go
type Plugin struct {
    // Manifest
    Name            string
    Version         string
    Description     string
    Author          string
    
    // Filesystem
    Path            string          // Plugin root directory
    EntryPoint      string          // Main executable or script
    
    // Configuration
    Config          PluginConfig
    Dependencies    map[string]string  // {name: version}
    
    // Tools
    Tools           []PluginTool
    
    // Lifecycle
    IsLoaded        bool
    LoadedAt        time.Time
    Process         *os.Process     // For external plugins
}

type PluginTool struct {
    Name            string
    Description     string
    InputSchema     map[string]interface{}
    OutputSchema    map[string]interface{}
    
    // Handler
    Handler         func(input interface{}) (interface{}, error)
    IsAsync         bool
}

type PluginConfig struct {
    Enabled         bool
    Permissions     []string        // "file_read", "network", etc.
    EnvVars         map[string]string
    ResourceLimits  ResourceLimits
    Timeout         time.Duration
}
```

#### Plugin Manifest (plugin.yaml)

```yaml
name: my-plugin
version: 1.0.0
description: My custom plugin
author: John Doe

dependencies:
  base-plugin: ">=1.0.0"
  helper-lib: "^2.0.0"

tools:
  - name: analyze-code
    description: Analyze code quality
    input_schema:
      type: object
      properties:
        code:
          type: string
        language:
          type: string
    output_schema:
      type: object
      properties:
        issues:
          type: array
        score:
          type: number

permissions:
  - file_read
  - network_http

config:
  timeout: 30s
  env_vars:
    PLUGIN_DEBUG: "false"
```

---

## 3. Integration Points

### 3.1 Phase 2 → Phase 3 Integration

```
Phase 2 Components          Phase 3 Extensions
─────────────────          ──────────────────

Agent (Discovery)    ←────→  Docker Executor (Sandbox)
       │                         │
       └────→ Run Command  ─────→┤
                                 │
                            Docker Config
                                 │
                            Resource Limits
                                 │
                            Volume Mounts

Dependency Graph     ←────→  MCP Tool Discovery
       │                     Plugin Tools
       └─────────────────────────┘

Session              ←────→  Checkpointing
       │                    Context Compression
       └─────────────────────────┘

Tool Registry        ←────→  MCP Client
       │                     Plugin Registry
       ├─────────────────────────┘
       │
       └───→ Batch Executor
```

### 3.2 ADK Framework Integration

```
ADK FunctionTool Pattern (Phase 2)
        ↓
┌──────────────────────────────────────┐
│ New Phase 3 Tools                    │
│                                      │
│ 1. sandbox_run                       │
│    Input: SandboxRunInput            │
│    Output: SandboxRunOutput          │
│    Handler: DockerExecutor.Run()     │
│                                      │
│ 2. list_mcp_tools                    │
│    Input: MCPToolsInput              │
│    Output: MCPToolsList              │
│    Handler: MCPClient.DiscoverTools()│
│                                      │
│ 3. load_plugin                       │
│    Input: LoadPluginInput            │
│    Output: LoadPluginOutput          │
│    Handler: PluginLoader.Load()      │
│                                      │
│ 4. batch_execute                     │
│    Input: BatchExecuteInput          │
│    Output: []BatchOutput             │
│    Handler: BatchExecutor.Execute()  │
└──────────────────────────────────────┘
```

---

## 4. Data Flow Diagrams

### 4.1 Docker Execution Flow

```
Tool Invocation (ADK)
         │
         ↓
┌─────────────────────┐
│ ExecutionRouter     │
│ Select strategy     │
└────────┬────────────┘
         │
         ├─ Strategy = "docker" ?
         │       │
         │       ↓
         │   ┌─────────────────────────┐
         │   │ DockerExecutor.Execute()│
         │   │                         │
         │   │ 1. Prepare container    │
         │   │    • Image pull         │
         │   │    • Env injection      │
         │   │    • Volume mount       │
         │   │                         │
         │   │ 2. Run container        │
         │   │    • Start              │
         │   │    • Stream output      │
         │   │    • Monitor resources  │
         │   │                         │
         │   │ 3. Capture result       │
         │   │    • Exit code          │
         │   │    • Stdout/stderr      │
         │   │    • Timing             │
         │   │                         │
         │   │ 4. Cleanup              │
         │   │    • Stop container     │
         │   │    • Remove container   │
         │   │    • Log audit          │
         │   └────────┬────────────────┘
         │            │
         └────┬───────┘
              │
              ↓
     ┌────────────────────┐
     │ ExecutionResult    │
     │ • Status           │
     │ • Output           │
     │ • ExitCode         │
     │ • Duration         │
     │ • Metrics          │
     └────────────────────┘
```

### 4.2 MCP Tool Integration Flow

```
Agent requests tool "analyze-code"
         │
         ↓
┌─────────────────────────────┐
│ Tool Registry lookup        │
│ • Check standard tools      │
│ • Check MCP tools           │
│ • Check plugin tools        │
└────────┬────────────────────┘
         │
         ├─ Found in MCP?
         │       │
         │       ↓
         │   ┌──────────────────────────────┐
         │   │ MCPClient.CallTool()         │
         │   │                              │
         │   │ 1. Validate input schema     │
         │   │ 2. Send to MCP server        │
         │   │ 3. Wait for result           │
         │   │ 4. Validate output schema    │
         │   │ 5. Return to agent           │
         │   └──────────────────────────────┘
         │
         └────┬───────────────────┐
              │                   │
              ↓                   ↓
     Tool Result           Fallback
     (MCP)                 (not found)
```

### 4.3 Batch Execution Flow

```
Batch Input (JSON)
         │
         ├─ Query: "Refactor function"
         ├─ Params: {file: "main.go"}
         └─ Timeout: 60
         │
         ↓
┌──────────────────────────────┐
│ BatchExecutor.Execute()      │
│                              │
│ 1. Validate inputs           │
│ 2. Setup environment         │
│ 3. For each input:           │
│    ├─ Validate parameters    │
│    ├─ Call orchestrator      │
│    ├─ Capture result         │
│    └─ Format output          │
│ 4. Aggregate results         │
└──────────┬───────────────────┘
           │
           ├─ Success → JSON output
           ├─ Error → Error JSON + exit 1
           └─ Timeout → Timeout JSON + exit 124
```

---

## 5. Error Handling & Recovery

### 5.1 Docker-Specific Errors

```go
type DockerError struct {
    Code        int
    Message     string
    Recoverable bool
    Suggestion  string
}

var DockerErrors = map[int]DockerError{
    1: {
        Code: 1,
        Message: "Image not found",
        Recoverable: true,
        Suggestion: "Run: docker pull <image>",
    },
    2: {
        Code: 2,
        Message: "Container startup failed",
        Recoverable: true,
        Suggestion: "Check Docker daemon, disk space, or resource limits",
    },
    3: {
        Code: 3,
        Message: "Command timeout",
        Recoverable: false,
        Suggestion: "Increase timeout or optimize command",
    },
    4: {
        Code: 4,
        Message: "Resource limit exceeded",
        Recoverable: false,
        Suggestion: "Increase memory/CPU limits or optimize command",
    },
}
```

### 5.2 MCP Connection Recovery

```go
func (c *MCPClient) withRetry(fn func() error) error {
    for attempt := 0; attempt < c.RetryPolicy.MaxRetries; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        // Check if recoverable
        if isTemporaryError(err) {
            backoff := exponentialBackoff(attempt, c.RetryPolicy.BaseDelay)
            time.Sleep(backoff)
            continue
        }
        
        return err
    }
    return ErrMaxRetriesExceeded
}
```

---

## 6. Security Considerations

### 6.1 Docker Security

```
┌─────────────────────────────────┐
│ Security Layers                 │
├─────────────────────────────────┤
│ 1. Image Verification           │
│    └─ Only trusted images       │
│                                 │
│ 2. Resource Limits              │
│    ├─ Memory (prevent OOM)      │
│    ├─ CPU (prevent hijack)      │
│    └─ Disk (prevent filling)    │
│                                 │
│ 3. Network Isolation            │
│    ├─ No network by default     │
│    └─ Explicit allow only       │
│                                 │
│ 4. Volume Restrictions          │
│    ├─ Read-only mounts          │
│    └─ Path validation           │
│                                 │
│ 5. Credential Injection         │
│    ├─ Masked in logs            │
│    ├─ Not in filesystem         │
│    └─ Securely removed          │
│                                 │
│ 6. Audit Logging                │
│    ├─ All commands logged       │
│    ├─ All outputs captured      │
│    └─ Tamper detection          │
└─────────────────────────────────┘
```

### 6.2 Plugin Security Model

```
Plugin Manifest → Permission Checker → Approval
    │                   │
    ├─ Permissions      ├─ file_read?
    ├─ Resource limits  ├─ network?
    └─ Dependencies     ├─ subprocess?
                        └─ User approval?
    │
    ↓
    Sandboxed Execution
    • Limited filesystem access
    • Limited network access
    • Resource limits enforced
    • Process isolation
```

---

## 7. Configuration Management

### 7.1 Phase 3 Configuration Schema

```yaml
# .adk/config.yaml (extended from Phase 2)

# Phase 2 configs still work...
agent:
  skip_missing: false

search_order:
  - project
  - user
  - plugin

# NEW: Phase 3 configs
execution:
  strategy: "docker"  # or "direct", "ssh"
  
docker:
  image: "golang:1.24"
  image_pull_policy: "IfNotPresent"
  
  resources:
    memory_mb: 512
    cpu_shares: 1024
    timeout_seconds: 300
  
  volumes:
    - host_path: "."
      container_path: "/workspace"
      read_only: false

credentials:
  sources:
    - type: "environment"
      prefix: "ADK_"
    - type: "vault"
      address: "https://vault.example.com"
      token_env: "VAULT_TOKEN"

audit:
  enabled: true
  log_file: "~/.adk/audit.json"
  include_output: true
  log_level: "info"

headless:
  output_format: "json"
  continue_on_error: false
  max_retries: 3

mcp:
  servers:
    - name: "my-server"
      type: "stdio"
      command: "python3 /path/to/server.py"

plugins:
  enabled: true
  search_paths:
    - "~/.adk/plugins"
    - "./.adk/plugins"
  
  permissions:
    default: "deny"  # Explicit allow required
    trust_signed: true
```

---

## 8. Testing Strategy

### 8.1 Unit Test Coverage

```
Docker Executor (40+ tests)
├─ Container creation
├─ Resource limit enforcement
├─ Volume mounting
├─ Environment variables
├─ Timeout handling
├─ Error scenarios

MCP Client (35+ tests)
├─ Connection/disconnection
├─ Tool discovery
├─ Tool invocation
├─ Error handling
├─ Retry logic

Plugin System (30+ tests)
├─ Plugin loading
├─ Permission checking
├─ Tool registration
├─ Dependency resolution

Batch Executor (25+ tests)
├─ Input validation
├─ Output formatting
├─ Error handling
├─ Timeout behavior

Session Checkpoint (20+ tests)
├─ Compression triggers
├─ Message summarization
├─ Recovery from checkpoint
```

### 8.2 Integration Tests

```
E2E Docker Execution (10+ tests)
├─ Run simple command
├─ Run with volumes
├─ Run with env vars
├─ Handle timeout
├─ Handle errors
├─ Cleanup verification

E2E MCP Integration (8+ tests)
├─ Connect to server
├─ Discover tools
├─ Call tool successfully
├─ Handle server errors
├─ Reconnect after failure

E2E Plugin Loading (6+ tests)
├─ Load valid plugin
├─ Load invalid plugin
├─ Use plugin tool
├─ Handle plugin errors

E2E Batch Execution (5+ tests)
├─ Execute batch jobs
├─ Resume from checkpoint
├─ Handle mixed success/failure
```

---

## 9. Performance Targets

### 9.1 Latency Goals

| Operation | Target | Current | Status |
|-----------|--------|---------|--------|
| Docker container start | <2s | - | 📋 TBD |
| MCP tool discovery | <100ms | - | 📋 TBD |
| Plugin loading | <500ms | - | 📋 TBD |
| Batch job (small) | <5s | - | 📋 TBD |
| Context compression | <1s | - | 📋 TBD |

### 9.2 Resource Goals

| Resource | Limit | Monitoring |
|----------|-------|-----------|
| Docker container memory | Configurable | OS metrics |
| Plugin process memory | <100MB | Process metrics |
| Audit log size | <10GB | File rotation |
| MCP connection pool | 10 concurrent | Connection counter |

---

## 10. Deployment & Operations

### 10.1 Prerequisite Installation

```bash
# Docker (required for Docker sandboxing)
brew install docker

# Go modules (automatically handled)
go mod tidy

# Optional: Docker images
docker pull golang:1.24
docker pull python:3.11
docker pull node:20
```

### 10.2 Health Check Script

```bash
#!/bin/bash
# Check Phase 3 components

echo "🔍 Checking Docker..."
docker info > /dev/null && echo "✅ Docker OK" || echo "❌ Docker failed"

echo "🔍 Checking adk-code binary..."
./adk-code --version > /dev/null && echo "✅ Binary OK" || echo "❌ Binary failed"

echo "🔍 Testing Docker sandboxing..."
./adk-code /sandbox "echo 'test'" && echo "✅ Sandboxing OK" || echo "❌ Sandboxing failed"

echo "🔍 Testing batch mode..."
echo '{"query":"test"}' | ./adk-code --headless --json && echo "✅ Batch mode OK" || echo "❌ Batch mode failed"
```

---

## Conclusion

Phase 3 Architecture provides a **layered, extensible, secure foundation** for production deployment of adk-code. By separating concerns (Docker execution, credential management, plugins, MCP) into independent subsystems, it maintains code clarity while enabling sophisticated features.

**Key Architecture Benefits**:
1. **Backward Compatible**: Phase 2 code works unchanged
2. **Pluggable**: Easy to swap execution strategies
3. **Observable**: Complete audit trail
4. **Secure**: Defense-in-depth approach
5. **Scalable**: Can support multiple execution backends

---

**Next Steps**: Proceed to implementation (Phase 3.1) with Docker Sandboxing as first priority.
