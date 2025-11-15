// Package web provides web content fetching tools for the coding agent.
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	common "adk-code/tools/base"
)

// FetchWebInput defines parameters for fetching web content.
type FetchWebInput struct {
	// URL to fetch (required)
	URL string `json:"url" jsonschema:"URL to fetch (e.g., https://example.com/page)"`

	// Format specifies how to process the response (optional)
	// "text" (default) - plain text extraction
	// "json" - parse as JSON
	// "html" - parse HTML structure
	// "raw" - return raw response
	Format *string `json:"format,omitempty" jsonschema:"Response format: 'text', 'json', 'html', 'raw' (default: text)"`

	// Timeout in seconds (optional, default: 30s)
	Timeout *int `json:"timeout,omitempty" jsonschema:"Request timeout in seconds (default: 30)"`

	// FollowRedirects controls automatic redirect following (optional, default: true)
	FollowRedirects *bool `json:"follow_redirects,omitempty" jsonschema:"Follow HTTP redirects (default: true)"`

	// MaxSize is the maximum response size in bytes (optional, default: 1MB)
	// Prevents fetching extremely large files
	MaxSize *int64 `json:"max_size,omitempty" jsonschema:"Maximum response size in bytes (default: 1048576)"`

	// Headers are optional custom HTTP headers to send with the request
	Headers map[string]string `json:"headers,omitempty" jsonschema:"Custom HTTP headers (e.g., Authorization)"`
}

// FetchWebOutput contains the fetched web content and metadata.
type FetchWebOutput struct {
	// Success indicates whether the fetch was successful
	Success bool `json:"success"`

	// Content is the fetched and optionally processed content
	Content string `json:"content"`

	// URL is the final URL after any redirects
	URL string `json:"url"`

	// StatusCode is the HTTP status code (e.g., 200, 404, 500)
	StatusCode int `json:"status_code"`

	// ContentType is the MIME type of the response (e.g., text/html, application/json)
	ContentType string `json:"content_type"`

	// ContentLength is the size of the response in bytes
	ContentLength int64 `json:"content_length"`

	// Headers contains response headers (optional, common ones only)
	Headers map[string]string `json:"headers,omitempty"`

	// ProcessedFormat indicates how the content was processed
	ProcessedFormat string `json:"processed_format"`

	// TruncatedAt indicates if content was truncated at this byte position
	TruncatedAt int64 `json:"truncated_at,omitempty"`

	// Error contains error message if the fetch failed
	Error string `json:"error,omitempty"`

	// ErrorCode provides a machine-readable error classification
	// "network_error", "timeout", "status_error", "too_large", "parsing_error", etc.
	ErrorCode string `json:"error_code,omitempty"`

	// FetchDurationMS is the time taken to fetch in milliseconds
	FetchDurationMS int `json:"fetch_duration_ms"`
}

// FetchWebHandler implements the web fetch logic.
func FetchWebHandler(ctx tool.Context, input FetchWebInput) FetchWebOutput {
	startTime := time.Now()
	output := FetchWebOutput{
		Success:         false,
		ProcessedFormat: getFormat(input.Format),
	}

	// 1. Validate URL
	parsedURL, err := url.Parse(input.URL)
	if err != nil {
		output.Error = fmt.Sprintf("Invalid URL: %v", err)
		output.ErrorCode = "invalid_url"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// Only allow HTTP and HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		output.Error = fmt.Sprintf("Unsupported URL scheme: %s (only http and https are supported)", parsedURL.Scheme)
		output.ErrorCode = "invalid_url"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// 2. Configure request
	client := &http.Client{
		Timeout:       getTimeout(input.Timeout),
		CheckRedirect: getRedirectPolicy(input.FollowRedirects),
	}

	// Use context.Background() if ctx is nil (for testing)
	// ctx implements both tool.Context and context.Context interfaces
	var reqCtx context.Context
	if ctx == nil {
		reqCtx = context.Background()
	} else {
		reqCtx = ctx
	}

	req, err := http.NewRequestWithContext(reqCtx, "GET", input.URL, nil)
	if err != nil {
		output.Error = fmt.Sprintf("Failed to create request: %v", err)
		output.ErrorCode = "request_error"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// 3. Add custom headers
	addHeaders(req, input.Headers)

	// 4. Execute request with timeout
	resp, err := client.Do(req)
	if err != nil {
		if os.IsTimeout(err) || err == context.DeadlineExceeded {
			output.Error = "Request timeout"
			output.ErrorCode = "timeout"
		} else {
			output.Error = fmt.Sprintf("Failed to fetch: %v", err)
			output.ErrorCode = "network_error"
		}
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}
	defer resp.Body.Close()

	// 5. Populate response metadata
	output.StatusCode = resp.StatusCode
	output.URL = resp.Request.URL.String()
	output.ContentType = resp.Header.Get("Content-Type")

	// 6. Check status code
	if resp.StatusCode >= 400 {
		output.Error = fmt.Sprintf("HTTP error %d", resp.StatusCode)
		output.ErrorCode = "status_error"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// 7. Check content size
	maxSize := getMaxSize(input.MaxSize)
	if resp.ContentLength > maxSize && resp.ContentLength > 0 {
		output.Error = fmt.Sprintf("Response too large: %d > %d bytes", resp.ContentLength, maxSize)
		output.ErrorCode = "too_large"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// 8. Read response with limit
	limitedReader := io.LimitReader(resp.Body, maxSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		output.Error = fmt.Sprintf("Failed to read response: %v", err)
		output.ErrorCode = "read_error"
		output.FetchDurationMS = int(time.Since(startTime).Milliseconds())
		return output
	}

	// Check if content was truncated
	if int64(len(content)) > maxSize {
		output.TruncatedAt = maxSize
		content = content[:maxSize]
	}

	// 9. Process content based on format
	processed, wasProcessed := processContent(
		string(content),
		resp.Header.Get("Content-Type"),
		getFormat(input.Format),
	)

	output.Success = true
	output.Content = processed
	output.ContentLength = int64(len(content))
	output.FetchDurationMS = int(time.Since(startTime).Milliseconds())

	if wasProcessed {
		output.ProcessedFormat = getFormat(input.Format)
	} else {
		output.ProcessedFormat = "raw"
	}

	return output
}

// getTimeout converts optional timeout in seconds to time.Duration.
func getTimeout(timeoutSeconds *int) time.Duration {
	const defaultTimeout = 30 * time.Second
	const maxTimeout = 5 * time.Minute

	if timeoutSeconds == nil {
		return defaultTimeout
	}

	duration := time.Duration(*timeoutSeconds) * time.Second
	if duration > maxTimeout {
		duration = maxTimeout
	}
	if duration <= 0 {
		duration = defaultTimeout
	}
	return duration
}

// getMaxSize returns the configured max response size with bounds checking.
func getMaxSize(maxSize *int64) int64 {
	const defaultMaxSize = 1024 * 1024     // 1 MB
	const absMaxSize = 50 * 1024 * 1024    // 50 MB hard limit

	if maxSize == nil {
		return defaultMaxSize
	}

	if *maxSize > absMaxSize {
		return absMaxSize
	}
	if *maxSize <= 0 {
		return defaultMaxSize
	}
	return *maxSize
}

// getFormat normalizes format string.
func getFormat(format *string) string {
	if format == nil || *format == "" {
		return "text"
	}
	f := strings.ToLower(*format)
	switch f {
	case "text", "json", "html", "raw":
		return f
	default:
		return "text"
	}
}

// getRedirectPolicy returns appropriate redirect policy function.
func getRedirectPolicy(followRedirects *bool) func(*http.Request, []*http.Request) error {
	follow := true
	if followRedirects != nil {
		follow = *followRedirects
	}

	if !follow {
		return func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return nil // Use default behavior
}

// addHeaders adds optional custom headers to the request.
func addHeaders(req *http.Request, headers map[string]string) {
	// Set User-Agent if not provided
	if headers == nil || headers["User-Agent"] == "" {
		req.Header.Set("User-Agent", "adk-code/1.0 (+https://github.com/raphaelmansuy/adk-code)")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// processContent parses and formats response based on requested format.
func processContent(content, contentType, format string) (string, bool) {
	switch format {
	case "json":
		return extractJSON(content, contentType)
	case "html":
		return extractHTML(content, contentType)
	case "raw":
		return content, false
	case "text":
		fallthrough
	default:
		return extractText(content, contentType)
	}
}

// extractText removes HTML tags and returns clean text.
func extractText(content, contentType string) (string, bool) {
	if !isHTMLContent(contentType) {
		return content, false
	}

	// Simple HTML tag removal using regexp
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(content, "")

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	return text, true
}

// extractHTML parses and returns HTML structure.
func extractHTML(content, contentType string) (string, bool) {
	if !isHTMLContent(contentType) {
		return content, false
	}

	// Parse HTML and extract main content
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return content, false
	}

	// Extract text and structure
	extracted := extractHTMLStructure(doc)
	return extracted, true
}

// extractHTMLStructure walks the HTML tree and extracts structured content.
func extractHTMLStructure(n *html.Node) string {
	var buf strings.Builder

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text)
				buf.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	result := buf.String()
	result = strings.TrimSpace(result)
	// Clean up multiple spaces
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	return result
}

// extractJSON validates and formats JSON.
func extractJSON(content, contentType string) (string, bool) {
	if !isJSONContent(contentType) {
		// Try to parse anyway in case content-type is wrong
	}

	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return content, false
	}

	// Re-marshal with indentation for readability
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return content, false
	}

	return string(pretty), true
}

// isHTMLContent checks if content-type indicates HTML.
func isHTMLContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html")
}

// isJSONContent checks if content-type indicates JSON.
func isJSONContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "text/json")
}

// NewFetchWebTool creates a tool for fetching web content.
func NewFetchWebTool() (tool.Tool, error) {
	t, err := functiontool.New(functiontool.Config{
		Name: "builtin_fetch_web",
		Description: `Fetches content from a web URL with optional parsing and formatting.

**Parameters:**
- url (required): The URL to fetch (http or https only)
- format (optional): How to process the response - "text" (default, extracts plain text from HTML), "json" (formats JSON), "html" (extracts HTML structure), "raw" (returns raw content)
- timeout (optional): Request timeout in seconds (default: 30, max: 300)
- follow_redirects (optional): Follow HTTP redirects (default: true)
- max_size (optional): Maximum response size in bytes (default: 1MB, max: 50MB)
- headers (optional): Custom HTTP headers as key-value pairs

**Use Cases:**
- Fetch documentation from specific URLs (README files, API docs, tutorials)
- Retrieve real-time data from web services and APIs
- Access web page content directly without searching
- Get JSON data from API endpoints
- Parse HTML content from web pages

**Examples:**
- Fetch README: url="https://raw.githubusercontent.com/google/adk-go/main/README.md", format="text"
- Fetch API data: url="https://api.example.com/data", format="json"
- Fetch web page: url="https://golang.org/doc/", format="text"

**Complementary to Google Search:**
- Use Google Search to find relevant URLs based on a query
- Use this tool to fetch content from specific URLs you already know

**Security:** Only HTTP/HTTPS protocols are supported. Response size and timeout limits prevent abuse.`,
	}, FetchWebHandler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  1, // Secondary to Google Search (priority 0)
			UsageHint: "Fetch content from specific URLs. Supports text, JSON, HTML parsing. Use after Google Search to retrieve content from found URLs.",
		})
	}

	return t, err
}
