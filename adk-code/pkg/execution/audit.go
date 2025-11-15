package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// AuditEventTypeCommandStart is recorded when a command starts
	AuditEventTypeCommandStart AuditEventType = "command_start"

	// AuditEventTypeCommandOutput is recorded when command produces output
	AuditEventTypeCommandOutput AuditEventType = "command_output"

	// AuditEventTypeCommandError is recorded when command fails
	AuditEventTypeCommandError AuditEventType = "command_error"

	// AuditEventTypeCommandEnd is recorded when a command completes
	AuditEventTypeCommandEnd AuditEventType = "command_end"

	// AuditEventTypeSecretAccess is recorded when a secret is accessed
	AuditEventTypeSecretAccess AuditEventType = "secret_access"

	// AuditEventTypeSecretCreate is recorded when a secret is created
	AuditEventTypeSecretCreate AuditEventType = "secret_create"

	// AuditEventTypeSecretDelete is recorded when a secret is deleted
	AuditEventTypeSecretDelete AuditEventType = "secret_delete"
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	// ID is a unique event identifier
	ID string `json:"id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Type is the event type
	Type AuditEventType `json:"type"`

	// ExecutionID is the execution context ID
	ExecutionID string `json:"execution_id,omitempty"`

	// Message is a human-readable event message
	Message string `json:"message"`

	// Details contains additional event data
	Details map[string]interface{} `json:"details,omitempty"`

	// Masked indicates whether sensitive data was masked
	Masked bool `json:"masked,omitempty"`

	// Source indicates where the event came from (agent, executor, etc.)
	Source string `json:"source,omitempty"`

	// UserID is the user who triggered the event
	UserID string `json:"user_id,omitempty"`

	// Error contains error information if applicable
	Error string `json:"error,omitempty"`

	// Duration contains operation duration in milliseconds
	Duration int64 `json:"duration_ms,omitempty"`
}

// AuditLog stores audit events
type AuditLog struct {
	// Events stores all audit events
	events []AuditEvent

	// mu protects concurrent access
	mu sync.RWMutex
}

// NewAuditLog creates a new audit log
func NewAuditLog() *AuditLog {
	return &AuditLog{
		events: make([]AuditEvent, 0),
	}
}

// Log adds an audit event to the log
func (al *AuditLog) Log(event AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	al.events = append(al.events, event)
}

// LogCommand logs a command execution event
func (al *AuditLog) LogCommand(executionID, command string, args []string, masked bool) {
	event := AuditEvent{
		ID:          fmt.Sprintf("cmd-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        AuditEventTypeCommandStart,
		ExecutionID: executionID,
		Message:     fmt.Sprintf("Command started: %s", command),
		Masked:      masked,
		Details: map[string]interface{}{
			"command": command,
			"args":    args,
		},
	}

	al.Log(event)
}

// LogOutput logs command output
func (al *AuditLog) LogOutput(executionID, output string, masked bool) {
	event := AuditEvent{
		ID:          fmt.Sprintf("out-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        AuditEventTypeCommandOutput,
		ExecutionID: executionID,
		Message:     fmt.Sprintf("Command output: %d bytes", len(output)),
		Masked:      masked,
		Details: map[string]interface{}{
			"output_length": len(output),
		},
	}

	al.Log(event)
}

// LogError logs a command error
func (al *AuditLog) LogError(executionID string, err error, exitCode int) {
	event := AuditEvent{
		ID:          fmt.Sprintf("err-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        AuditEventTypeCommandError,
		ExecutionID: executionID,
		Message:     fmt.Sprintf("Command error: %v (exit code: %d)", err, exitCode),
		Error:       err.Error(),
		Details: map[string]interface{}{
			"exit_code": exitCode,
		},
	}

	al.Log(event)
}

// LogSecretAccess logs secret access
func (al *AuditLog) LogSecretAccess(executionID, secretName string) {
	event := AuditEvent{
		ID:          fmt.Sprintf("secret-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        AuditEventTypeSecretAccess,
		ExecutionID: executionID,
		Message:     fmt.Sprintf("Secret accessed: %s", secretName),
		Masked:      true,
		Details: map[string]interface{}{
			"secret_name": secretName,
		},
	}

	al.Log(event)
}

// GetEvents returns all events with optional filtering
func (al *AuditLog) GetEvents() []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Return a copy to prevent external modification
	events := make([]AuditEvent, len(al.events))
	copy(events, al.events)

	return events
}

// GetEventsByType returns events of a specific type
func (al *AuditLog) GetEventsByType(eventType AuditEventType) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filtered []AuditEvent
	for _, event := range al.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByExecutionID returns events for a specific execution
func (al *AuditLog) GetEventsByExecutionID(executionID string) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filtered []AuditEvent
	for _, event := range al.events {
		if event.ExecutionID == executionID {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsSince returns events after a specific time
func (al *AuditLog) GetEventsSince(since time.Time) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filtered []AuditEvent
	for _, event := range al.events {
		if event.Timestamp.After(since) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// Count returns the total number of events
func (al *AuditLog) Count() int {
	al.mu.RLock()
	defer al.mu.RUnlock()

	return len(al.events)
}

// Clear removes all events
func (al *AuditLog) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.events = make([]AuditEvent, 0)
}

// ExportJSON exports all events as JSON
func (al *AuditLog) ExportJSON() ([]byte, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	return json.MarshalIndent(al.events, "", "  ")
}

// AuditLogger is a comprehensive audit logging system
type AuditLogger struct {
	log               *AuditLog
	credentialMgr     *CredentialManager
	maskSensitiveData bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(credentialMgr *CredentialManager, maskSensitive bool) *AuditLogger {
	return &AuditLogger{
		log:               NewAuditLog(),
		credentialMgr:     credentialMgr,
		maskSensitiveData: maskSensitive,
	}
}

// LogExecution logs a complete execution with timing
func (al *AuditLogger) LogExecution(ctx context.Context, executionID string, command string, args []string, output string, exitCode int, duration time.Duration) error {
	// Log command start
	al.log.LogCommand(executionID, command, args, al.maskSensitiveData)

	// Log output (may be masked)
	maskedOutput := output
	if al.maskSensitiveData && al.credentialMgr != nil {
		var err error
		maskedOutput, err = al.credentialMgr.MaskOutput(ctx, output)
		if err != nil {
			return err
		}
	}

	al.log.LogOutput(executionID, maskedOutput, al.maskSensitiveData)

	// Log error if non-zero exit code
	if exitCode != 0 {
		al.log.LogError(executionID, fmt.Errorf("command failed"), exitCode)
	}

	// Log completion
	event := AuditEvent{
		ID:          fmt.Sprintf("end-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        AuditEventTypeCommandEnd,
		ExecutionID: executionID,
		Message:     fmt.Sprintf("Command completed with exit code %d", exitCode),
		Duration:    duration.Milliseconds(),
	}

	al.log.Log(event)

	return nil
}

// GetLog returns the underlying audit log
func (al *AuditLogger) GetLog() *AuditLog {
	return al.log
}

// Summary generates a summary of logged events
func (al *AuditLogger) Summary(executionID string) map[string]interface{} {
	events := al.log.GetEventsByExecutionID(executionID)

	summary := map[string]interface{}{
		"execution_id":   executionID,
		"event_count":    len(events),
		"events_by_type": map[string]int{},
		"first_event":    nil,
		"last_event":     nil,
	}

	typeCounts := make(map[AuditEventType]int)

	if len(events) > 0 {
		summary["first_event"] = events[0].Timestamp
		summary["last_event"] = events[len(events)-1].Timestamp

		for _, event := range events {
			typeCounts[event.Type]++
		}
	}

	for eventType, count := range typeCounts {
		summary["events_by_type"].(map[string]int)[string(eventType)] = count
	}

	return summary
}
