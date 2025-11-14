package execution

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// DockerContainerConfig holds configuration for Docker container execution
type DockerContainerConfig struct {
	Image                    string
	Tag                      string
	Command                  []string
	WorkDir                  string
	EnvironmentVariables     map[string]string
	VolumeMounts             map[string]string
	MemoryLimit              int64
	CPUShares                int64
	Timeout                  time.Duration
	NetworkMode              string
	Privileged               bool
	RemoveAfterExecution     bool
}

// DockerContainerResult holds the result of Docker container execution
type DockerContainerResult struct {
	ContainerID   string
	Output        string
	Error         string
	ExitCode      int
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	ResourceUsage ResourceUsage
}

// ResourceUsage contains resource utilization information
type ResourceUsage struct {
	MemoryUsedMB    int64
	CPUTimeMS       int64
	NetworkBytesIn  int64
	NetworkBytesOut int64
}

// DockerExecutor manages Docker container execution using the Docker CLI
type DockerExecutor struct {
	// Placeholder for future enhancements (socket connection, etc.)
}

// NewDockerExecutor creates a new Docker executor using the Docker CLI
func NewDockerExecutor() (*DockerExecutor, error) {
	// Check if Docker is available
	if err := CheckDockerAvailable(); err != nil {
		return nil, err
	}
	return &DockerExecutor{}, nil
}

// Close closes the Docker executor (no-op for CLI-based executor)
func (de *DockerExecutor) Close() error {
	return nil
}

// Ping checks if Docker daemon is available
func (de *DockerExecutor) Ping(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "ps")
	return cmd.Run()
}

// PullImage pulls a Docker image from the registry using docker cli
func (de *DockerExecutor) PullImage(ctx context.Context, imageName string) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", imageName)
	return cmd.Run()
}

// Execute runs a command in a Docker container
func (de *DockerExecutor) Execute(ctx context.Context, config DockerContainerConfig) (*DockerContainerResult, error) {
	result := &DockerContainerResult{
		StartTime: time.Now(),
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Build docker run command
	args := []string{"run"}

	// Add removal flag
	if config.RemoveAfterExecution {
		args = append(args, "--rm")
	}

	// Add memory limit
	if config.MemoryLimit > 0 {
		args = append(args, "-m", fmt.Sprintf("%dM", config.MemoryLimit))
	}

	// Add CPU limit
	if config.CPUShares > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%d", config.CPUShares))
	}

	// Add network mode
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}

	// Add privileged mode
	if config.Privileged {
		args = append(args, "--privileged")
	}

	// Add working directory
	if config.WorkDir != "" {
		args = append(args, "-w", config.WorkDir)
	}

	// Add volume mounts
	for hostPath, containerPath := range config.VolumeMounts {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// Add environment variables
	for k, v := range config.EnvironmentVariables {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add image name
	imageName := config.Image
	if config.Tag != "" {
		imageName = fmt.Sprintf("%s:%s", config.Image, config.Tag)
	}
	args = append(args, imageName)

	// Add command and arguments
	args = append(args, config.Command...)

	// Execute the docker run command
	cmd := exec.CommandContext(execCtx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Output = stdout.String()
	if stderr.Len() > 0 {
		if result.Output != "" {
			result.Output += "\n"
		}
		result.Output += stderr.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Sprintf("exit code %d", result.ExitCode)
		} else {
			result.Error = err.Error()
		}
	} else {
		result.ExitCode = 0
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// GetContainerStats retrieves resource usage statistics (placeholder for now)
func (de *DockerExecutor) GetContainerStats(ctx context.Context, containerID string) (*ResourceUsage, error) {
	return &ResourceUsage{}, nil
}

// CheckDockerAvailable checks if Docker is installed and accessible
func CheckDockerAvailable() error {
	cmd := exec.Command("docker", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker is not available or not installed: %w", err)
	}
	return nil
}
