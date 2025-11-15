# Spec 0002: Memory & Artifact Interfaces

**Status**: Ready for Implementation  
**Priority**: P1  
**Effort**: 4 hours  
**Dependencies**: Spec 0001  
**Files**: `pkg/memory/memory.go`, `pkg/artifact/artifact.go`  

## Summary

Create Memory and Artifact service interfaces with no-op implementations for Phase 2, enabling Phase 3 real implementations without API changes.

## Memory Interface

**Location**: `pkg/memory/memory.go`

```go
// Memory defines semantic memory storage and search
type Memory interface {
    Save(ctx context.Context, content string, metadata map[string]interface{}) error
    Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
    Get(ctx context.Context, id string) (string, error)
    Delete(ctx context.Context, id string) error
}

// SearchResult represents a single memory search result
type SearchResult struct {
    ID       string                 // Memory ID
    Content  string                 // Memory content
    Score    float32                // Relevance score (0.0-1.0)
    Metadata map[string]interface{} // Original metadata
}

// DefaultMemory returns no-op Memory implementation
func DefaultMemory() Memory {
    return &noopMemory{}
}
```

## Artifact Interface

**Location**: `pkg/artifact/artifact.go`

```go
// Service defines artifact storage lifecycle
type Service interface {
    Save(ctx context.Context, artifact *Artifact) error
    Load(ctx context.Context, id string) (*Artifact, error)
    List(ctx context.Context) ([]*Artifact, error)
    Delete(ctx context.Context, id string) error
}

// Artifact represents a single file/document/result
type Artifact struct {
    ID       string                 // Unique ID
    Name     string                 // Human-readable name
    Type     string                 // "file", "document", "result"
    Content  []byte                 // File content
    Metadata map[string]interface{} // Custom metadata
}

// DefaultService returns no-op Artifact Service
func DefaultService() Service {
    return &noopService{}
}
```

## Implementation Steps

1. **Create Memory package**:
   - Define `Memory` interface
   - Define `SearchResult` struct
   - Implement `noopMemory` (no-op backend)
   - Implement `DefaultMemory()` function
   - Add unit tests

2. **Create Artifact package**:
   - Define `Service` interface
   - Define `Artifact` struct
   - Implement `noopService` (no-op backend)
   - Implement `DefaultService()` function
   - Add unit tests

3. **Export from execution.go**:
   - Add imports for Memory and Service
   - Add to ExecutionContext default constructors

## Testing

- Unit tests for no-op implementations
- Mock implementations for testing
- Integration tests with ExecutionContext

## Success Criteria

- [ ] Memory interface defined
- [ ] Artifact Service interface defined
- [ ] No-op implementations provided
- [ ] All unit tests pass
- [ ] Constructors work with defaults
- [ ] Ready for Phase 3 implementations

## Notes for Phase 3

- Memory will implement semantic search via embeddings
- Artifact will support file system or cloud storage
- Version control for artifacts
- Metadata tagging and filtering

---

**Version**: 1.0  
**Updated**: November 15, 2025
