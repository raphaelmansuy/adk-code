package display

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// MessageDeduplicator prevents duplicate message rendering
type MessageDeduplicator struct {
	mu           sync.RWMutex
	seen         map[string]time.Time
	cleanupTimer *time.Timer
	stopCh       chan struct{}
}

// NewMessageDeduplicator creates a new message deduplicator
func NewMessageDeduplicator() *MessageDeduplicator {
	md := &MessageDeduplicator{
		seen:   make(map[string]time.Time),
		stopCh: make(chan struct{}),
	}

	// Start cleanup routine
	md.startCleanup()

	return md
}

// IsDuplicate checks if a message is a duplicate
// Returns true if the message was seen recently
func (md *MessageDeduplicator) IsDuplicate(content string) bool {
	hash := md.hash(content)

	md.mu.RLock()
	_, exists := md.seen[hash]
	md.mu.RUnlock()

	if exists {
		return true
	}

	// Mark as seen
	md.mu.Lock()
	md.seen[hash] = time.Now()
	md.mu.Unlock()

	return false
}

// hash generates a hash for content
func (md *MessageDeduplicator) hash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// startCleanup starts periodic cleanup of old entries
func (md *MessageDeduplicator) startCleanup() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				md.cleanup()
			case <-md.stopCh:
				return
			}
		}
	}()
}

// cleanup removes entries older than 1 minute
func (md *MessageDeduplicator) cleanup() {
	md.mu.Lock()
	defer md.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Minute)
	for hash, timestamp := range md.seen {
		if timestamp.Before(cutoff) {
			delete(md.seen, hash)
		}
	}
}

// Stop stops the deduplicator's cleanup routine
func (md *MessageDeduplicator) Stop() {
	close(md.stopCh)
}

// Clear clears all stored message hashes
func (md *MessageDeduplicator) Clear() {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.seen = make(map[string]time.Time)
}
