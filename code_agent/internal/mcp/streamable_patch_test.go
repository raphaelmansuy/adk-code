package mcp

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFilterSSEEvents(t *testing.T) {
	// Test data with ping events
	input := `event: message
data: {"jsonrpc":"2.0","id":1,"result":{"test":"value1"}}

event: ping
data: ping

event: message
data: {"jsonrpc":"2.0","id":2,"result":{"test":"value2"}}

event: ping
data: ping

`

	// Expected output (ping events removed)
	expected := `event: message
data: {"jsonrpc":"2.0","id":1,"result":{"test":"value1"}}

event: message
data: {"jsonrpc":"2.0","id":2,"result":{"test":"value2"}}

`

	// Create pipe and filter
	pr, pw := io.Pipe()
	src := io.NopCloser(strings.NewReader(input))

	// Run filter in background
	go filterSSEEvents(src, pw)

	// Read filtered output
	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	if err != nil {
		t.Fatalf("Error reading filtered output: %v", err)
	}

	got := buf.String()
	if got != expected {
		t.Errorf("Filter output mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}
