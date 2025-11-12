// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"testing"
)

func TestNewAgentError(t *testing.T) {
	err := New(CodeFileNotFound, "test message")
	if err.Code != CodeFileNotFound {
		t.Errorf("Expected code %s, got %s", CodeFileNotFound, err.Code)
	}
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}
	if err.Error() != "[FILE_NOT_FOUND] test message" {
		t.Errorf("Unexpected error string: %s", err.Error())
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(CodeExecution, "execution failed", originalErr)

	if err.Code != CodeExecution {
		t.Errorf("Expected code %s, got %s", CodeExecution, err.Code)
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped error not preserved")
	}
	if !errors.Is(err, originalErr) {
		t.Errorf("errors.Is failed for wrapped error")
	}
}

func TestWithContext(t *testing.T) {
	err := New(CodeInvalidInput, "invalid").
		WithContext("param", "value1").
		WithContext("code", "ABC123")

	if err.Context["param"] != "value1" {
		t.Errorf("Context not set correctly")
	}
	if err.Context["code"] != "ABC123" {
		t.Errorf("Context not set correctly")
	}
}

func TestIsFunction(t *testing.T) {
	err := New(CodeFileNotFound, "file not found")
	if !Is(err, CodeFileNotFound) {
		t.Errorf("Is() should return true for matching error code")
	}
	if Is(err, CodePermission) {
		t.Errorf("Is() should return false for non-matching error code")
	}

	// Test with non-AgentError
	genericErr := errors.New("generic error")
	if Is(genericErr, CodeFileNotFound) {
		t.Errorf("Is() should return false for non-AgentError")
	}
}

func TestFileNotFoundError(t *testing.T) {
	err := FileNotFoundError("/path/to/file.txt")
	if err.Code != CodeFileNotFound {
		t.Errorf("Wrong error code")
	}
	if err.Context["path"] != "/path/to/file.txt" {
		t.Errorf("Path not set in context")
	}
}

func TestPermissionDeniedError(t *testing.T) {
	err := PermissionDeniedError("/root/secret")
	if err.Code != CodePermission {
		t.Errorf("Wrong error code")
	}
	if err.Context["path"] != "/root/secret" {
		t.Errorf("Path not set in context")
	}
}

func TestPathTraversalError(t *testing.T) {
	err := PathTraversalError("/workspace/../../../etc/passwd", "/workspace")
	if err.Code != CodePathTraversal {
		t.Errorf("Wrong error code")
	}
	if err.Context["path"] != "/workspace/../../../etc/passwd" {
		t.Errorf("Path not set in context")
	}
	if err.Context["base_path"] != "/workspace" {
		t.Errorf("Base path not set in context")
	}
}

func TestSymlinkEscapeError(t *testing.T) {
	err := SymlinkEscapeError("/workspace/link", "/etc/passwd", "/workspace")
	if err.Code != CodeSymlinkEscape {
		t.Errorf("Wrong error code")
	}
	if err.Context["path"] != "/workspace/link" {
		t.Errorf("Path not set in context")
	}
	if err.Context["real_path"] != "/etc/passwd" {
		t.Errorf("Real path not set in context")
	}
	if err.Context["base_path"] != "/workspace" {
		t.Errorf("Base path not set in context")
	}
}

func TestExecutionError(t *testing.T) {
	innerErr := errors.New("command not found")
	err := ExecutionError("bash", innerErr)
	if err.Code != CodeExecution {
		t.Errorf("Wrong error code")
	}
	if err.Wrapped != innerErr {
		t.Errorf("Inner error not wrapped")
	}
	if err.Context["command"] != "bash" {
		t.Errorf("Command not set in context")
	}
}

func TestTimeoutError(t *testing.T) {
	err := TimeoutError("download")
	if err.Code != CodeTimeout {
		t.Errorf("Wrong error code")
	}
	if err.Context["operation"] != "download" {
		t.Errorf("Operation not set in context")
	}
}

func TestAPIKeyError(t *testing.T) {
	err := APIKeyError("openai")
	if err.Code != CodeAPIKey {
		t.Errorf("Wrong error code")
	}
	if err.Context["provider"] != "openai" {
		t.Errorf("Provider not set in context")
	}
}

func TestModelNotFoundError(t *testing.T) {
	err := ModelNotFoundError("claude-3")
	if err.Code != CodeModelNotFound {
		t.Errorf("Wrong error code")
	}
	if err.Context["model_id"] != "claude-3" {
		t.Errorf("Model ID not set in context")
	}
}

func TestProviderError(t *testing.T) {
	innerErr := errors.New("connection refused")
	err := ProviderError("gemini", innerErr)
	if err.Code != CodeProviderError {
		t.Errorf("Wrong error code")
	}
	if err.Context["provider"] != "gemini" {
		t.Errorf("Provider not set in context")
	}
	if err.Wrapped != innerErr {
		t.Errorf("Inner error not wrapped")
	}
}

func TestPatchFailedError(t *testing.T) {
	err := PatchFailedError("line 10 does not match")
	if err.Code != CodePatchFailed {
		t.Errorf("Wrong error code")
	}
}

func TestInternalError(t *testing.T) {
	err := InternalError("null pointer")
	if err.Code != CodeInternal {
		t.Errorf("Wrong error code")
	}
}

func TestNotSupportedError(t *testing.T) {
	err := NotSupportedError("streaming mode")
	if err.Code != CodeNotSupported {
		t.Errorf("Wrong error code")
	}
	if err.Context["feature"] != "streaming mode" {
		t.Errorf("Feature not set in context")
	}
}

func TestErrorStringFormat(t *testing.T) {
	tests := []struct {
		err      *AgentError
		expected string
	}{
		{
			New(CodeFileNotFound, "file missing"),
			"[FILE_NOT_FOUND] file missing",
		},
		{
			Wrap(CodeExecution, "failed", errors.New("timeout")),
			"[EXECUTION_FAILED] failed: timeout",
		},
	}

	for i, test := range tests {
		if test.err.Error() != test.expected {
			t.Errorf("Test %d: Expected '%s', got '%s'", i, test.expected, test.err.Error())
		}
	}
}
