package display

import (
	"fmt"
	"sync"
	"time"
)

// SpinnerStyle defines the animation style for a spinner
type SpinnerStyle struct {
	Frames []string
	Speed  time.Duration
}

// Predefined spinner styles
var (
	SpinnerDots = SpinnerStyle{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Speed:  80 * time.Millisecond,
	}

	SpinnerLine = SpinnerStyle{
		Frames: []string{"-", "\\", "|", "/"},
		Speed:  100 * time.Millisecond,
	}

	SpinnerArrow = SpinnerStyle{
		Frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		Speed:  100 * time.Millisecond,
	}

	SpinnerCircle = SpinnerStyle{
		Frames: []string{"◐", "◓", "◑", "◒"},
		Speed:  120 * time.Millisecond,
	}
)

// Spinner provides animated progress indication
type Spinner struct {
	mu       sync.Mutex
	style    SpinnerStyle
	message  string
	active   bool
	stopped  bool
	stopCh   chan struct{}
	doneCh   chan struct{}
	renderer *Renderer
}

// NewSpinner creates a new spinner
func NewSpinner(renderer *Renderer, message string) *Spinner {
	return &Spinner{
		style:    SpinnerDots,
		message:  message,
		renderer: renderer,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// NewSpinnerWithStyle creates a new spinner with a custom style
func NewSpinnerWithStyle(renderer *Renderer, message string, style SpinnerStyle) *Spinner {
	return &Spinner{
		style:    style,
		message:  message,
		renderer: renderer,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't start if already active or stopped
	if s.active || s.stopped {
		return
	}

	s.active = true
	s.stopped = false

	// Only show spinner in TTY mode
	if !IsTTY() {
		fmt.Printf("%s...\n", s.message)
		return
	}

	// Recreate channels for reuse
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go s.animate()
}

// Stop stops the spinner animation (safe to call multiple times)
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already stopped to avoid closing an already closed channel
	if !s.active || s.stopped {
		return
	}

	s.stopped = true
	s.active = false
	close(s.stopCh)

	// Release lock before waiting for doneCh to avoid deadlock
	s.mu.Unlock()
	<-s.doneCh // Wait for animation to stop
	s.mu.Lock()
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.style.Speed)
	defer ticker.Stop()

	frame := 0

	for {
		select {
		case <-s.stopCh:
			// Clear the line
			fmt.Print("\r\033[K")
			return

		case <-ticker.C:
			s.mu.Lock()
			message := s.message
			s.mu.Unlock()

			// Render current frame
			spinChar := s.style.Frames[frame%len(s.style.Frames)]

			// Use renderer colors if available
			var output string
			if s.renderer != nil {
				output = fmt.Sprintf("\r%s %s",
					s.renderer.Cyan(spinChar),
					s.renderer.Dim(message))
			} else {
				output = fmt.Sprintf("\r%s %s", spinChar, message)
			}

			fmt.Print(output)

			frame++
		}
	}
}

// StopWithMessage stops the spinner and prints a final message
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	if message != "" {
		fmt.Println(message)
	}
}

// StopWithSuccess stops the spinner and prints a success message
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	if s.renderer != nil {
		fmt.Println(s.renderer.SuccessCheckmark(message))
	} else {
		fmt.Printf("✓ %s\n", message)
	}
}

// StopWithError stops the spinner and prints an error message
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	if s.renderer != nil {
		fmt.Println(s.renderer.ErrorX(message))
	} else {
		fmt.Printf("✗ %s\n", message)
	}
}
