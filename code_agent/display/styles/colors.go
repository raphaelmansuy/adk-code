package styles

import "github.com/charmbracelet/lipgloss"

// Styles holds all the lipgloss style definitions
type Styles struct {
	DimStyle     lipgloss.Style
	GreenStyle   lipgloss.Style
	RedStyle     lipgloss.Style
	YellowStyle  lipgloss.Style
	BlueStyle    lipgloss.Style
	CyanStyle    lipgloss.Style
	WhiteStyle   lipgloss.Style
	BoldStyle    lipgloss.Style
	SuccessStyle lipgloss.Style
}

// NewStyles initializes all lipgloss styles
func NewStyles() *Styles {
	return &Styles{
		DimStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),            // Bright black (gray)
		GreenStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),            // Green
		RedStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("1")),            // Red
		YellowStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),            // Yellow
		BlueStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("39")),           // Bright blue
		CyanStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("6")),            // Cyan
		WhiteStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("7")),            // White
		BoldStyle:    lipgloss.NewStyle().Bold(true),                                 // Bold
		SuccessStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true), // Green + Bold
	}
}
