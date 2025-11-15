package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchWebTool_Basic(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	input := FetchWebInput{
		URL: server.URL,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if !output.Success {
		t.Errorf("Expected success, got error: %s", output.Error)
	}
	if output.Content != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got: %s", output.Content)
	}
	if output.StatusCode != 200 {
		t.Errorf("Expected status code 200, got: %d", output.StatusCode)
	}
	if output.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got: %s", output.ContentType)
	}
}

func TestFetchWebTool_InvalidURL(t *testing.T) {
	input := FetchWebInput{
		URL: "not-a-valid-url",
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if output.Success {
		t.Error("Expected failure for invalid URL")
	}
	if output.ErrorCode != "invalid_url" {
		t.Errorf("Expected error code 'invalid_url', got: %s", output.ErrorCode)
	}
}

func TestFetchWebTool_UnsupportedScheme(t *testing.T) {
	input := FetchWebInput{
		URL: "ftp://example.com/file.txt",
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if output.Success {
		t.Error("Expected failure for unsupported scheme")
	}
	if output.ErrorCode != "invalid_url" {
		t.Errorf("Expected error code 'invalid_url', got: %s", output.ErrorCode)
	}
	if !strings.Contains(output.Error, "Unsupported URL scheme") {
		t.Errorf("Expected unsupported scheme error, got: %s", output.Error)
	}
}

func TestFetchWebTool_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	input := FetchWebInput{
		URL: server.URL,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if output.Success {
		t.Error("Expected failure for 404 error")
	}
	if output.ErrorCode != "status_error" {
		t.Errorf("Expected error code 'status_error', got: %s", output.ErrorCode)
	}
	if output.StatusCode != 404 {
		t.Errorf("Expected status code 404, got: %d", output.StatusCode)
	}
}

func TestFetchWebTool_ResponseTooLarge(t *testing.T) {
	// Create a test server with known large content
	largeContent := strings.Repeat("X", 2*1024*1024) // 2 MB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", "2097152") // 2 MB
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	maxSize := int64(1024 * 1024) // 1 MB limit
	input := FetchWebInput{
		URL:     server.URL,
		MaxSize: &maxSize,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if output.Success {
		t.Error("Expected failure for response too large")
	}
	if output.ErrorCode != "too_large" {
		t.Errorf("Expected error code 'too_large', got: %s", output.ErrorCode)
	}
}

func TestFetchWebTool_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Delayed response"))
	}))
	defer server.Close()

	timeout := 1 // 1 second timeout
	input := FetchWebInput{
		URL:     server.URL,
		Timeout: &timeout,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if output.Success {
		t.Error("Expected failure due to timeout")
	}
	if output.ErrorCode != "timeout" && output.ErrorCode != "network_error" {
		t.Errorf("Expected error code 'timeout' or 'network_error', got: %s", output.ErrorCode)
	}
}

func TestFetchWebTool_Redirects(t *testing.T) {
	// Create a test server with redirect
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final destination"))
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	input := FetchWebInput{
		URL: redirectServer.URL,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if !output.Success {
		t.Errorf("Expected success, got error: %s", output.Error)
	}
	if output.Content != "Final destination" {
		t.Errorf("Expected content 'Final destination', got: %s", output.Content)
	}
	if output.URL == redirectServer.URL {
		t.Error("Expected final URL to be different from redirect URL")
	}
}

func TestFetchWebTool_NoRedirects(t *testing.T) {
	// Create a test server with redirect
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Should not see this"))
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	followRedirects := false
	input := FetchWebInput{
		URL:             redirectServer.URL,
		FollowRedirects: &followRedirects,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	// When not following redirects, we get a 301 response which is technically successful
	// (status code < 400), so we check that we got the redirect response itself
	if !output.Success {
		t.Errorf("Expected success (even without following redirect), got error: %s", output.Error)
	}
	if output.StatusCode != 301 {
		t.Errorf("Expected status code 301, got: %d", output.StatusCode)
	}
	// Verify we got the redirect URL, not the final content
	if output.URL == finalServer.URL {
		t.Error("Should not have followed redirect to final URL")
	}
}

func TestExtractText_RemovesHTMLTags(t *testing.T) {
	html := "<html><body><h1>Title</h1><p>Paragraph</p></body></html>"
	result, wasProcessed := extractText(html, "text/html")

	if !wasProcessed {
		t.Error("Expected HTML to be processed")
	}
	if strings.Contains(result, "<") || strings.Contains(result, ">") {
		t.Errorf("Expected HTML tags to be removed, got: %s", result)
	}
	if !strings.Contains(result, "Title") || !strings.Contains(result, "Paragraph") {
		t.Errorf("Expected text content to be preserved, got: %s", result)
	}
}

func TestExtractText_PlainText(t *testing.T) {
	text := "Plain text content"
	result, wasProcessed := extractText(text, "text/plain")

	if wasProcessed {
		t.Error("Expected plain text to not be processed")
	}
	if result != text {
		t.Errorf("Expected text to be unchanged, got: %s", result)
	}
}

func TestExtractJSON_FormatsJSON(t *testing.T) {
	jsonStr := `{"name":"test","value":123}`
	result, wasProcessed := extractJSON(jsonStr, "application/json")

	if !wasProcessed {
		t.Error("Expected JSON to be processed")
	}
	if !strings.Contains(result, "\n") {
		t.Error("Expected formatted JSON with newlines")
	}
	if !strings.Contains(result, "test") || !strings.Contains(result, "123") {
		t.Errorf("Expected JSON content to be preserved, got: %s", result)
	}
}

func TestExtractJSON_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`
	result, wasProcessed := extractJSON(invalidJSON, "application/json")

	if wasProcessed {
		t.Error("Expected invalid JSON to not be processed")
	}
	if result != invalidJSON {
		t.Errorf("Expected original content to be returned, got: %s", result)
	}
}

func TestGetTimeout_EnforcesLimits(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected time.Duration
	}{
		{"nil (default)", nil, 30 * time.Second},
		{"valid timeout", intPtr(60), 60 * time.Second},
		{"exceeds max", intPtr(600), 5 * time.Minute},
		{"zero", intPtr(0), 30 * time.Second},
		{"negative", intPtr(-10), 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTimeout(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetMaxSize_EnforcesLimits(t *testing.T) {
	tests := []struct {
		name     string
		input    *int64
		expected int64
	}{
		{"nil (default)", nil, 1024 * 1024},
		{"valid size", int64Ptr(2 * 1024 * 1024), 2 * 1024 * 1024},
		{"exceeds max", int64Ptr(100 * 1024 * 1024), 50 * 1024 * 1024},
		{"zero", int64Ptr(0), 1024 * 1024},
		{"negative", int64Ptr(-1000), 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxSize(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetFormat_Normalization(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, "text"},
		{strPtr(""), "text"},
		{strPtr("text"), "text"},
		{strPtr("TEXT"), "text"},
		{strPtr("json"), "json"},
		{strPtr("JSON"), "json"},
		{strPtr("html"), "html"},
		{strPtr("raw"), "raw"},
		{strPtr("invalid"), "text"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%v", tt.input), func(t *testing.T) {
			result := getFormat(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFetchWebTool_CustomHeaders(t *testing.T) {
	// Create a test server that checks headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	input := FetchWebInput{
		URL: server.URL,
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if !output.Success {
		t.Errorf("Expected success, got error: %s", output.Error)
	}
	if output.Content != "Authenticated" {
		t.Errorf("Expected content 'Authenticated', got: %s", output.Content)
	}
}

func TestFetchWebTool_JSONFormat(t *testing.T) {
	jsonData := `{"message": "Hello", "count": 42}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonData))
	}))
	defer server.Close()

	format := "json"
	input := FetchWebInput{
		URL:    server.URL,
		Format: &format,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if !output.Success {
		t.Errorf("Expected success, got error: %s", output.Error)
	}
	if output.ProcessedFormat != "json" {
		t.Errorf("Expected processed format 'json', got: %s", output.ProcessedFormat)
	}
	if !strings.Contains(output.Content, "Hello") || !strings.Contains(output.Content, "42") {
		t.Errorf("Expected JSON content to be preserved, got: %s", output.Content)
	}
}

func TestFetchWebTool_HTMLFormat(t *testing.T) {
	htmlContent := `<html><head><title>Test</title></head><body><h1>Header</h1><p>Content</p></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	format := "html"
	input := FetchWebInput{
		URL:    server.URL,
		Format: &format,
	}

	// Use nil context for unit testing
	output := FetchWebHandler(nil, input)

	if !output.Success {
		t.Errorf("Expected success, got error: %s", output.Error)
	}
	if output.ProcessedFormat != "html" {
		t.Errorf("Expected processed format 'html', got: %s", output.ProcessedFormat)
	}
	if !strings.Contains(output.Content, "Header") || !strings.Contains(output.Content, "Content") {
		t.Errorf("Expected HTML text to be extracted, got: %s", output.Content)
	}
}

func TestNewFetchWebTool_CreatesValidTool(t *testing.T) {
	tool, err := NewFetchWebTool()
	if err != nil {
		t.Fatalf("Failed to create fetch web tool: %v", err)
	}
	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
