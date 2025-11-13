package mcp

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"
)

// This file contains a workaround for MCP SDK bug #636:
// https://github.com/modelcontextprotocol/go-sdk/issues/636
//
// The SDK tries to JSON-decode ALL SSE events, including "ping" events.
// This wrapper filters out non-"message" events before they reach the SDK.
//
// TODO: Remove this file once the SDK is fixed (likely v1.2.0)

// filteringTransport wraps an http.RoundTripper to filter SSE ping events
type filteringTransport struct {
	base http.RoundTripper
}

func (t *filteringTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Only filter text/event-stream responses
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		return resp, nil
	}

	// Create a pipe for streaming filtered events
	pr, pw := io.Pipe()

	// Start filtering in background
	go filterSSEEvents(resp.Body, pw)

	// Replace response body with filtered pipe reader
	resp.Body = pr

	return resp, nil
}

// filterSSEEvents reads from src, filters out non-message events, and writes to dst
func filterSSEEvents(src io.ReadCloser, dst *io.PipeWriter) {
	defer src.Close()
	defer dst.Close()

	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 0, 64*1024), 1*1024*1024)

	var eventBuffer bytes.Buffer
	var eventName string

	for scanner.Scan() {
		line := scanner.Text() // Use Text() instead of Bytes() to get a copy

		// Empty line = end of event
		if len(line) == 0 {
			if eventBuffer.Len() > 0 {
				// Include event if it's "message" or has no explicit name
				if eventName == "" || eventName == "message" {
					// Write event to output
					if _, err := dst.Write(eventBuffer.Bytes()); err != nil {
						return
					}
					if _, err := dst.Write([]byte("\n")); err != nil {
						return
					}
				}
				// Reset for next event
				eventBuffer.Reset()
				eventName = ""
			}
			continue
		}

		// Check if this is an event: line
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(line[6:])
		}

		// Accumulate event lines
		eventBuffer.WriteString(line)
		eventBuffer.WriteString("\n")
	}

	// Handle final event if stream ended without trailing newline
	if eventBuffer.Len() > 0 {
		if eventName == "" || eventName == "message" {
			dst.Write(eventBuffer.Bytes())
		}
	}

	// Note: scanner errors are expected when the context is canceled
	// The SDK cancels the context after reading the data it needs
	if err := scanner.Err(); err != nil {
		dst.CloseWithError(err)
	}
}
