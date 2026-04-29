package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

var (
	modalBox = lipgloss.NewStyle().
			Border(theme.Border).
			BorderForeground(theme.ColorAccent).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a1a"))

	modalTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorAccent).
			Margin(0, 0, 1, 0)
)

// Modal renderiza title + body em uma caixa centralizada. Quando width ou
// height são 0, retorna apenas a caixa (sem Place wrapper) — útil para
// inspeção em testes ou composições adicionais. Quando width > 0 e height > 0,
// a caixa é centralizada num canvas dessa dimensão.
func Modal(title, body string, width, height int) string {
	box := modalBox.Render(modalTitle.Render(title) + "\n" + body)
	if width <= 0 || height <= 0 {
		return box
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
