package styles

// IsTTY is imported from display package for formatting decisions
var IsTTY func() bool

// SetTTYCheck allows display package to inject IsTTY function
func SetTTYCheck(fn func() bool) {
	IsTTY = fn
}

// Formatting constants
const (
	OutputFormatRich  = "rich"
	OutputFormatPlain = "plain"
	OutputFormatJSON  = "json"
)

// Formatter provides text formatting methods
type Formatter struct {
	styles       *Styles
	outputFormat string
}

// NewFormatter creates a new formatter with the given output format
func NewFormatter(outputFormat string, styles *Styles) *Formatter {
	return &Formatter{
		styles:       styles,
		outputFormat: outputFormat,
	}
}

// shouldFormat returns true if formatting should be applied
func (f *Formatter) shouldFormat() bool {
	return f.outputFormat != OutputFormatPlain && IsTTY != nil && IsTTY()
}

// Dim renders text in dim gray
func (f *Formatter) Dim(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.DimStyle.Render(text)
}

// Green renders text in green
func (f *Formatter) Green(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.GreenStyle.Render(text)
}

// Red renders text in red
func (f *Formatter) Red(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.RedStyle.Render(text)
}

// Yellow renders text in yellow
func (f *Formatter) Yellow(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.YellowStyle.Render(text)
}

// Blue renders text in blue
func (f *Formatter) Blue(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.BlueStyle.Render(text)
}

// Cyan renders text in cyan
func (f *Formatter) Cyan(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.CyanStyle.Render(text)
}

// Bold renders text in bold
func (f *Formatter) Bold(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.BoldStyle.Render(text)
}

// Success renders text in green with bold
func (f *Formatter) Success(text string) string {
	if !f.shouldFormat() {
		return text
	}
	return f.styles.SuccessStyle.Render(text)
}

// SuccessCheckmark renders a checkmark with text in green
func (f *Formatter) SuccessCheckmark(text string) string {
	return f.Success("✓ " + text)
}

// ErrorX renders an X with text in red
func (f *Formatter) ErrorX(text string) string {
	return f.Red("✗ " + text)
}
