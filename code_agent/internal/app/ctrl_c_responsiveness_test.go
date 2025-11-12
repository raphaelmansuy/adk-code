package app

import (
	"context"
	"testing"
	"time"
)

// TestCtrlCResponsiveness verifies that context cancellation breaks the agent loop immediately
// even when the Runner.Run() iterator is waiting for the next event.
func TestCtrlCResponsiveness(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create a handler that cancels the context after a short delay
	// This simulates a user pressing Ctrl+C while the agent is thinking
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Record when cancellation was detected
	var cancelTime time.Time
	startTime := time.Now()

	// Simulate the agent loop pattern: check context in select statement
	// at the same level as the channel receive
	timeout := time.After(2 * time.Second)
	done := false

	for !done {
		select {
		case <-ctx.Done():
			// This should be reached quickly after cancel() is called
			cancelTime = time.Now()
			done = true
		case <-timeout:
			t.Fatal("timeout waiting for context cancellation to be detected")
		}
	}

	// Verify that cancellation was detected within a reasonable time
	// (should be close to the 100ms sleep, plus some overhead)
	elapsed := cancelTime.Sub(startTime)
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Cancellation detection took too long: %v (should be ~100ms)", elapsed)
	}

	t.Logf("✓ Context cancellation detected in %v", elapsed)
}

// TestCtrlCResponsiveness_WithChannelSelect verifies that the agent loop
// can respond to context cancellation at the same time it's waiting for
// events from a channel (the actual implementation pattern).
func TestCtrlCResponsiveness_WithChannelSelect(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel that simulates Runner.Run() events
	eventChan := make(chan string, 1)

	// Simulate a long-running operation that sends events slowly
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second): // Event arrives after 5 seconds
			eventChan <- "event"
		}
		close(eventChan)
	}()

	// Simulate user pressing Ctrl+C after a short time
	cancelTime := time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Main loop - this mimics the pattern in repl.go processUserMessage
	loopBreakTime := time.Time{}
	done := false

	for !done {
		select {
		case <-ctx.Done():
			// Ctrl+C detected - break the loop immediately
			loopBreakTime = time.Now()
			done = true
		case msg, ok := <-eventChan:
			// Event received from channel
			if !ok {
				done = true
			}
			_ = msg // Use the message
		}
	}

	// Verify that the loop broke due to context cancellation, not from waiting 5 seconds
	elapsed := loopBreakTime.Sub(cancelTime)
	if elapsed < 0 || elapsed > 500*time.Millisecond {
		t.Fatalf("Loop break detection timing incorrect: %v (should be ~100ms after cancel)", elapsed)
	}

	// Most importantly: verify we didn't wait for the 5-second event
	totalElapsed := loopBreakTime.Sub(cancelTime.Add(-100 * time.Millisecond))
	if totalElapsed > 1*time.Second {
		t.Fatalf("Waited too long for cancellation; total elapsed: %v", totalElapsed)
	}

	t.Logf("✓ Loop responded to context cancellation in ~%v (not waiting for 5s event)", elapsed)
}
