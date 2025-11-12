package components

import (
	"fmt"
	"sync"
	"time"

	"code_agent/internal/display/core"
	"code_agent/tracking"
)

// SpinnerStyle defines the animation style for a spinner
type SpinnerStyle struct {
	Frames []string
	Speed  time.Duration
}

// SpinnerMode defines the operational mode of the spinner
type SpinnerMode string

// Spinner modes
const (
	SpinnerModeTool     SpinnerMode = "tool"
	SpinnerModeThinking SpinnerMode = "thinking"
	SpinnerModeProgress SpinnerMode = "progress"
)

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

	// Thinking mode uses slower animation to convey deliberation
	SpinnerThinking = SpinnerStyle{
		Frames: []string{"◜", "◠", "◝", "◞", "◡", "◟"},
		Speed:  150 * time.Millisecond, // Slower than default
	}
)

// Spinner provides animated progress indication
type Spinner struct {
	mu       sync.Mutex
	style    SpinnerStyle
	message  string
	metrics  *tracking.TokenMetrics
	active   bool
	stopped  bool
	stopCh   chan struct{}
	doneCh   chan struct{}
	renderer core.StyleRenderer
	mode     SpinnerMode
}

// NewSpinner creates a new spinner
func NewSpinner(renderer core.StyleRenderer, message string) *Spinner {
	return &Spinner{
		style:    SpinnerDots,
		message:  message,
		renderer: renderer,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		mode:     SpinnerModeTool,
	}
}

// NewSpinnerWithStyle creates a new spinner with a custom style
func NewSpinnerWithStyle(renderer core.StyleRenderer, message string, style SpinnerStyle) *Spinner {
	return &Spinner{
		style:    style,
		message:  message,
		renderer: renderer,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		mode:     SpinnerModeTool,
	}
}

// SetMode sets the spinner mode (tool, thinking, or progress)
func (s *Spinner) SetMode(mode SpinnerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode

	// Update style based on mode
	switch mode {
	case SpinnerModeThinking:
		s.style = SpinnerThinking
	case SpinnerModeProgress:
		s.style = SpinnerDots // Could be different in future
	default:
		s.style = SpinnerDots
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't start if already active
	if s.active {
		return
	}

	// Reset stopped flag to allow restart
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
	s.metrics = nil
}

// UpdateWithMetrics updates the spinner message and token metrics
func (s *Spinner) UpdateWithMetrics(message string, metrics *tracking.TokenMetrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
	s.metrics = metrics
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
			// Print final state with metrics if available before clearing
			s.mu.Lock()
			message := s.message
			metrics := s.metrics
			mode := s.mode
			s.mu.Unlock()

			if metrics != nil && metrics.TotalTokens > 0 {
				metricsStr := tracking.FormatTokenMetrics(*metrics)
				if s.renderer != nil {
					var coloredSpinChar string
					if mode == SpinnerModeThinking {
						coloredSpinChar = s.renderer.Yellow(s.style.Frames[frame%len(s.style.Frames)])
					} else {
						coloredSpinChar = s.renderer.Cyan(s.style.Frames[frame%len(s.style.Frames)])
					}
					fmt.Print("\r" + coloredSpinChar + " " +
						s.renderer.Dim(message) + "  " + s.renderer.Dim(metricsStr) + "\n")
				} else {
					fmt.Printf("\r%s %s  %s\n", s.style.Frames[frame%len(s.style.Frames)], message, metricsStr)
				}
			} else {
				// No metrics, just clear
				fmt.Print("\r\033[K")
			}
			return

		case <-ticker.C:
			s.mu.Lock()
			message := s.message
			metrics := s.metrics
			mode := s.mode
			s.mu.Unlock()

			// Render current frame
			spinChar := s.style.Frames[frame%len(s.style.Frames)]

			// Build output with metrics if available
			var output string
			if metrics != nil && metrics.TotalTokens > 0 {
				metricsStr := tracking.FormatTokenMetrics(*metrics)
				if s.renderer != nil {
					// Color based on mode
					var coloredSpinChar string
					if mode == SpinnerModeThinking {
						coloredSpinChar = s.renderer.Yellow(spinChar)
					} else {
						coloredSpinChar = s.renderer.Cyan(spinChar)
					}
					output = fmt.Sprintf("\r%s %s  %s",
						coloredSpinChar,
						s.renderer.Dim(message),
						s.renderer.Dim(metricsStr))
				} else {
					output = fmt.Sprintf("\r%s %s  %s", spinChar, message, metricsStr)
				}
			} else {
				// No metrics, just show message
				if s.renderer != nil {
					// Color based on mode
					var coloredSpinChar string
					if mode == SpinnerModeThinking {
						coloredSpinChar = s.renderer.Yellow(spinChar)
					} else {
						coloredSpinChar = s.renderer.Cyan(spinChar)
					}
					output = fmt.Sprintf("\r%s %s",
						coloredSpinChar,
						s.renderer.Dim(message))
				} else {
					output = fmt.Sprintf("\r%s %s", spinChar, message)
				}
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
