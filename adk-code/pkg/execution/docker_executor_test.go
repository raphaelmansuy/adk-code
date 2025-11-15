package execution

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestNewDockerExecutor verifies Docker executor creation
func TestNewDockerExecutor(t *testing.T) {
	executor, err := NewDockerExecutor()

	// If Docker is not available, skip the test
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	if executor == nil {
		t.Fatal("Expected executor to be non-nil")
	}
}

// TestDockerExecutorPing verifies Docker daemon connectivity
func TestDockerExecutorPing(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = executor.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to ping Docker daemon: %v", err)
	}
}

// TestDockerExecutorPullImage verifies image pulling
func TestDockerExecutorPullImage(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use a small, commonly available image
	err = executor.PullImage(ctx, "alpine:latest")
	if err != nil {
		t.Fatalf("Failed to pull image: %v", err)
	}
}

// TestDockerExecutorExecuteBasic verifies basic command execution
func TestDockerExecutorExecuteBasic(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"echo", "hello"},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Fatalf("Expected output containing 'hello', got: %s", result.Output)
	}
}

// TestDockerExecutorExecuteWithMemoryLimit verifies memory limit enforcement
func TestDockerExecutorExecuteWithMemoryLimit(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"echo", "memory test"},
		MemoryLimit:          512,
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}
}

// TestDockerExecutorExecuteWithEnvironment verifies environment variables
func TestDockerExecutorExecuteWithEnvironment(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo $TEST_VAR"},
		EnvironmentVariables: map[string]string{
			"TEST_VAR": "test_value",
		},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "test_value") {
		t.Fatalf("Expected output containing 'test_value', got: %s", result.Output)
	}
}

// TestDockerExecutorExecuteTimeout verifies timeout handling
func TestDockerExecutorExecuteTimeout(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"sleep", "10"},
		Timeout:              1 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Should timeout before completing
	if result.ExitCode == 0 {
		t.Fatalf("Expected non-zero exit code due to timeout, got %d", result.ExitCode)
	}
}

// TestDockerContainerConfigValid verifies valid configuration
func TestDockerContainerConfigValid(t *testing.T) {
	config := DockerContainerConfig{
		Image:                "golang:1.24",
		Tag:                  "latest",
		Command:              []string{"go", "version"},
		WorkDir:              "/app",
		MemoryLimit:          1024,
		Timeout:              30 * time.Second,
		RemoveAfterExecution: true,
	}

	if config.Image == "" {
		t.Fatal("Image should not be empty")
	}

	if config.Timeout == 0 {
		t.Fatal("Timeout should be set")
	}
}

// TestDockerContainerResultTiming verifies timing information
func TestDockerContainerResultTiming(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"echo", "timing test"},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result.StartTime.IsZero() {
		t.Fatal("StartTime should be set")
	}

	if result.EndTime.IsZero() {
		t.Fatal("EndTime should be set")
	}

	if result.Duration == 0 {
		t.Fatal("Duration should be non-zero")
	}

	if result.EndTime.Before(result.StartTime) {
		t.Fatalf("EndTime (%v) should be after StartTime (%v)", result.EndTime, result.StartTime)
	}
}

// TestCheckDockerAvailable verifies Docker availability check
func TestCheckDockerAvailable(t *testing.T) {
	err := CheckDockerAvailable()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
}

// TestDockerExecutorClose verifies executor cleanup
func TestDockerExecutorClose(t *testing.T) {
	executor := &DockerExecutor{}
	err := executor.Close()
	if err != nil {
		t.Fatalf("Close() should not return error, got: %v", err)
	}
}

// TestDockerExecutorGetContainerStats verifies stats retrieval
func TestDockerExecutorGetContainerStats(t *testing.T) {
	executor := &DockerExecutor{}
	stats, err := executor.GetContainerStats(context.Background(), "test-container")
	if err != nil {
		t.Fatalf("GetContainerStats should not fail, got: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should be non-nil")
	}
}

// TestDockerExecutorExitCode verifies exit code handling
func TestDockerExecutorExitCode(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"sh", "-c", "exit 42"},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err == nil && result.ExitCode != 42 {
		t.Fatalf("Expected exit code 42, got %d", result.ExitCode)
	}
}

// TestDockerExecutorMultipleCommands verifies chained command execution
func TestDockerExecutorMultipleCommands(t *testing.T) {
	executor, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"sh", "-c", "echo hello && echo world"},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, config)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "hello") || !strings.Contains(result.Output, "world") {
		t.Fatalf("Expected output containing both 'hello' and 'world', got: %s", result.Output)
	}
}

// TestDockerConfigWithDefaults verifies default timeout
func TestDockerConfigWithDefaults(t *testing.T) {
	config := DockerContainerConfig{
		Image:   "alpine:latest",
		Command: []string{"echo", "test"},
	}

	if config.Timeout == 0 {
		// Timeout defaults to 30s in Execute method
		executor, err := NewDockerExecutor()
		if err != nil {
			t.Skipf("Docker not available: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, _ := executor.Execute(ctx, config)
		if result != nil && result.Duration == 0 {
			t.Fatal("Duration should be calculated even with default timeout")
		}
	}
}

// BenchmarkDockerExecute measures execution performance
func BenchmarkDockerExecute(b *testing.B) {
	executor, err := NewDockerExecutor()
	if err != nil {
		b.Skipf("Docker not available: %v", err)
	}

	config := DockerContainerConfig{
		Image:                "alpine:latest",
		Command:              []string{"echo", "bench"},
		Timeout:              5 * time.Second,
		RemoveAfterExecution: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, _ = executor.Execute(ctx, config)
		cancel()
	}
}
