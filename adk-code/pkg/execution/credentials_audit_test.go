package execution

import (
	"context"
	"testing"
	"time"
)

// TestSecret tests the Secret type
func TestSecret(t *testing.T) {
	secret := &Secret{
		Name:      "api_key",
		Value:     "secret123",
		Type:      "api_key",
		CreatedAt: time.Now(),
		Masked:    true,
	}

	if secret.Name != "api_key" {
		t.Fatalf("Expected name 'api_key', got %q", secret.Name)
	}

	if secret.IsExpired() {
		t.Fatal("Secret should not be expired")
	}
}

// TestSecretExpiry tests secret expiration
func TestSecretExpiry(t *testing.T) {
	secret := &Secret{
		Name:      "expired_key",
		Value:     "secret123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	if !secret.IsExpired() {
		t.Fatal("Secret should be expired")
	}
}

// TestSecretMaskedValue tests value masking
func TestSecretMaskedValue(t *testing.T) {
	secret := &Secret{
		Name:   "token",
		Value:  "super_secret_token",
		Masked: true,
	}

	masked := secret.MaskedValue()
	if masked == "super_secret_token" {
		t.Fatal("Expected value to be masked")
	}

	if masked == "****" {
		// Short values are fully masked
		t.Fatal("Expected partial masking for long values")
	}

	// Check that it starts and ends with real characters
	if masked[0] != 's' {
		t.Fatalf("Expected masked value to start with 's', got %c", masked[0])
	}
}

// TestInMemoryCredentialStoreStore tests storing credentials
func TestInMemoryCredentialStoreStore(t *testing.T) {
	store := NewInMemoryCredentialStore()
	ctx := context.Background()

	secret := &Secret{
		Name:  "test_key",
		Value: "test_value",
		Type:  "api_key",
	}

	err := store.Store(ctx, secret)
	if err != nil {
		t.Fatalf("Failed to store secret: %v", err)
	}
}

// TestInMemoryCredentialStoreRetrieve tests retrieving credentials
func TestInMemoryCredentialStoreRetrieve(t *testing.T) {
	store := NewInMemoryCredentialStore()
	ctx := context.Background()

	secret := &Secret{
		Name:  "test_key",
		Value: "test_value",
		Type:  "api_key",
	}

	err := store.Store(ctx, secret)
	if err != nil {
		t.Fatalf("Failed to store secret: %v", err)
	}

	retrieved, err := store.Retrieve(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to retrieve secret: %v", err)
	}

	if retrieved.Name != "test_key" {
		t.Fatalf("Expected name 'test_key', got %q", retrieved.Name)
	}

	if retrieved.Value != "test_value" {
		t.Fatalf("Expected value 'test_value', got %q", retrieved.Value)
	}
}

// TestInMemoryCredentialStoreList tests listing credentials
func TestInMemoryCredentialStoreList(t *testing.T) {
	store := NewInMemoryCredentialStore()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		secret := &Secret{
			Name:  "key" + string(rune(i)),
			Value: "value",
			Type:  "api_key",
		}
		_ = store.Store(ctx, secret)
	}

	names, err := store.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list secrets: %v", err)
	}

	if len(names) != 3 {
		t.Fatalf("Expected 3 secrets, got %d", len(names))
	}
}

// TestInMemoryCredentialStoreDelete tests deleting credentials
func TestInMemoryCredentialStoreDelete(t *testing.T) {
	store := NewInMemoryCredentialStore()
	ctx := context.Background()

	secret := &Secret{
		Name:  "test_key",
		Value: "test_value",
		Type:  "api_key",
	}

	_ = store.Store(ctx, secret)

	err := store.Delete(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	_, err = store.Retrieve(ctx, "test_key")
	if err == nil {
		t.Fatal("Expected error retrieving deleted secret")
	}
}

// TestCredentialManagerAddSecret tests adding secrets
func TestCredentialManagerAddSecret(t *testing.T) {
	manager := NewCredentialManager(nil)
	ctx := context.Background()

	err := manager.AddSecret(ctx, "api_key", "secret_value", "api_key")
	if err != nil {
		t.Fatalf("Failed to add secret: %v", err)
	}

	secret, err := manager.GetSecret(ctx, "api_key")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if secret.Value != "secret_value" {
		t.Fatalf("Expected value 'secret_value', got %q", secret.Value)
	}
}

// TestCredentialManagerGetSecretValue tests retrieving secret values
func TestCredentialManagerGetSecretValue(t *testing.T) {
	manager := NewCredentialManager(nil)
	ctx := context.Background()

	_ = manager.AddSecret(ctx, "api_key", "secret_value", "api_key")

	value, err := manager.GetSecretValue(ctx, "api_key")
	if err != nil {
		t.Fatalf("Failed to get secret value: %v", err)
	}

	if value != "secret_value" {
		t.Fatalf("Expected value 'secret_value', got %q", value)
	}
}

// TestCredentialManagerListSecrets tests listing secrets
func TestCredentialManagerListSecrets(t *testing.T) {
	manager := NewCredentialManager(nil)
	ctx := context.Background()

	_ = manager.AddSecret(ctx, "key1", "value1", "api_key")
	_ = manager.AddSecret(ctx, "key2", "value2", "token")

	names, err := manager.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("Failed to list secrets: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(names))
	}
}

// TestCredentialManagerMaskOutput tests output masking
func TestCredentialManagerMaskOutput(t *testing.T) {
	manager := NewCredentialManager(nil)
	ctx := context.Background()

	_ = manager.AddSecret(ctx, "api_key", "secret_token_123", "api_key")

	output := "Request with api_key: secret_token_123 completed"
	masked, err := manager.MaskOutput(ctx, output)
	if err != nil {
		t.Fatalf("Failed to mask output: %v", err)
	}

	if masked == output {
		t.Fatal("Expected output to be masked")
	}

	// Check that the secret was replaced (output should be shorter or different)
	if masked != output && len(masked) > 0 {
		// Output was successfully masked
		return
	}

	t.Fatalf("Output masking failed: %s", masked)
}

// TestAuditEventBasic tests basic audit event
func TestAuditEventBasic(t *testing.T) {
	event := AuditEvent{
		ID:        "test-1",
		Timestamp: time.Now(),
		Type:      AuditEventTypeCommandStart,
		Message:   "Command started",
	}

	if event.Type != AuditEventTypeCommandStart {
		t.Fatal("Expected AuditEventTypeCommandStart")
	}
}

// TestAuditLogLog tests logging events
func TestAuditLogLog(t *testing.T) {
	log := NewAuditLog()

	event := AuditEvent{
		ID:      "test-1",
		Type:    AuditEventTypeCommandStart,
		Message: "Command started",
	}

	log.Log(event)

	if log.Count() != 1 {
		t.Fatalf("Expected 1 event, got %d", log.Count())
	}
}

// TestAuditLogLogCommand tests logging commands
func TestAuditLogLogCommand(t *testing.T) {
	log := NewAuditLog()

	log.LogCommand("exec-1", "echo", []string{"hello"}, false)

	if log.Count() != 1 {
		t.Fatalf("Expected 1 event, got %d", log.Count())
	}

	events := log.GetEventsByType(AuditEventTypeCommandStart)
	if len(events) != 1 {
		t.Fatalf("Expected 1 command start event, got %d", len(events))
	}
}

// TestAuditLogLogOutput tests logging output
func TestAuditLogLogOutput(t *testing.T) {
	log := NewAuditLog()

	log.LogOutput("exec-1", "output data", false)

	events := log.GetEventsByType(AuditEventTypeCommandOutput)
	if len(events) != 1 {
		t.Fatalf("Expected 1 output event, got %d", len(events))
	}
}

// TestAuditLogGetEventsByExecutionID tests filtering by execution ID
func TestAuditLogGetEventsByExecutionID(t *testing.T) {
	log := NewAuditLog()

	log.LogCommand("exec-1", "cmd1", nil, false)
	log.LogCommand("exec-2", "cmd2", nil, false)

	events := log.GetEventsByExecutionID("exec-1")
	if len(events) != 1 {
		t.Fatalf("Expected 1 event for exec-1, got %d", len(events))
	}
}

// TestAuditLogGetEventsSince tests filtering by time
func TestAuditLogGetEventsSince(t *testing.T) {
	log := NewAuditLog()

	before := time.Now().Add(-1 * time.Minute)

	log.LogCommand("exec-1", "cmd1", nil, false)

	events := log.GetEventsSince(before)
	if len(events) != 1 {
		t.Fatalf("Expected 1 recent event, got %d", len(events))
	}
}

// TestAuditLogExportJSON tests JSON export
func TestAuditLogExportJSON(t *testing.T) {
	log := NewAuditLog()

	log.LogCommand("exec-1", "echo", []string{"test"}, false)

	json, err := log.ExportJSON()
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}

	if len(json) == 0 {
		t.Fatal("Expected non-empty JSON output")
	}
}

// TestAuditLoggerLogExecution tests comprehensive execution logging
func TestAuditLoggerLogExecution(t *testing.T) {
	manager := NewCredentialManager(nil)
	logger := NewAuditLogger(manager, true)

	ctx := context.Background()

	err := logger.LogExecution(ctx, "exec-1", "echo", []string{"hello"}, "hello\n", 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to log execution: %v", err)
	}

	if logger.GetLog().Count() < 2 {
		t.Fatalf("Expected at least 2 events logged, got %d", logger.GetLog().Count())
	}
}

// TestAuditLoggerSummary tests generating summary
func TestAuditLoggerSummary(t *testing.T) {
	logger := NewAuditLogger(nil, false)
	ctx := context.Background()

	_ = logger.LogExecution(ctx, "exec-1", "echo", []string{"test"}, "output", 0, 50*time.Millisecond)

	summary := logger.Summary("exec-1")
	if summary["execution_id"] != "exec-1" {
		t.Fatal("Expected correct execution_id in summary")
	}

	if eventCount, ok := summary["event_count"]; ok {
		count := eventCount.(int)
		if count < 2 {
			t.Fatalf("Expected at least 2 events in summary, got %d", count)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
