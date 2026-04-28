// Package theme exposes shared lipgloss styles.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (GitHub dark-ish palette).
	ColorAccent     = lipgloss.Color("#58a6ff")
	ColorOK         = lipgloss.Color("#3fb950")
	ColorWarn       = lipgloss.Color("#d29922")
	ColorError      = lipgloss.Color("#f85149")
	ColorDim        = lipgloss.Color("#6e7681")
	ColorSelectedBG = lipgloss.Color("#1f6feb")
	ColorSelectedFG = lipgloss.Color("#ffffff")

	// Borders & layout.
	Border = lipgloss.RoundedBorder()

	// Text styles.
	Title    = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
	Subtitle = lipgloss.NewStyle().Foreground(ColorDim)
	OK       = lipgloss.NewStyle().Foreground(ColorOK)
	Warn     = lipgloss.NewStyle().Foreground(ColorWarn)
	Error    = lipgloss.NewStyle().Foreground(ColorError)
	Selected = lipgloss.NewStyle().Background(ColorSelectedBG).Foreground(ColorSelectedFG)

	// Tab styles.
	TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSelectedFG).
			Background(ColorSelectedBG).
			Padding(0, 1)
	TabInactive = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)

	// Panes.
	Pane = lipgloss.NewStyle().
		Border(Border).
		BorderForeground(ColorDim).
		Padding(0, 1)
)
