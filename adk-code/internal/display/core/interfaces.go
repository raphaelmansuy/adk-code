package core

// StyleRenderer defines the interface for rendering styled text.
// This interface allows components to work with any renderer implementation
// without creating import cycles.
type StyleRenderer interface {
	// Color methods
	Yellow(text string) string
	Cyan(text string) string

	// Style methods
	Dim(text string) string

	// Status indicators
	SuccessCheckmark(text string) string
	ErrorX(text string) string
}
