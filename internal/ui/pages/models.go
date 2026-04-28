// Package pages holds tab page implementations.
package pages

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/modelscanner"
)

// pathStatus tracks per-root scan progress shown above the table.
type pathStatus struct {
	state string // "scanning" | "scanned" | "error"
	count int
	err   string
}

// ModelsPage browses GGUF files discovered by ModelScanner.
type ModelsPage struct {
	scanner modelscanner.Scanner
	paths   []string
	cancel  context.CancelFunc

	files     []domain.ModelFile
	statusMap map[string]pathStatus

	table      table.Model
	width      int
	height     int
	filter     string
	filterMode bool
	flash      string

	keys modelsKeyMap
}

type modelsKeyMap struct {
	Filter, Rescan, Enter, Cancel key.Binding
}

func defaultModelsKeys() modelsKeyMap {
	return modelsKeyMap{
		Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Rescan: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rescan")),
		Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "actions")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear filter")),
	}
}

// NewModelsPage builds a page wired to a Scanner and configured search paths.
func NewModelsPage(scanner modelscanner.Scanner, paths []string) ModelsPage {
	cols := []table.Column{
		{Title: "Name", Width: 36},
		{Title: "Size", Width: 10},
		{Title: "Quant", Width: 10},
		{Title: "Params", Width: 8},
		{Title: "Path", Width: 40},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true), table.WithHeight(12))

	statusMap := make(map[string]pathStatus, len(paths))
	for _, p := range paths {
		statusMap[p] = pathStatus{state: "scanning"}
	}
	return ModelsPage{
		scanner:   scanner,
		paths:     paths,
		statusMap: statusMap,
		table:     t,
		keys:      defaultModelsKeys(),
	}
}
