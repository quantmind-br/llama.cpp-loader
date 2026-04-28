// Command llama-cpp-loader launches the TUI for managing llama.cpp profiles.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/config"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/llamahelp"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/pages"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	store, err := profilestore.NewFSStore(cfg.Paths.ProfilesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "profile store: %v\n", err)
		os.Exit(1)
	}

	root := ui.NewRoot(parseTab(cfg.UI.DefaultTab)).
		WithProfilesPage(pages.NewProfilesPage(store, llamahelp.EmbeddedSchema()))

	prog := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}

func parseTab(name string) ui.Tab {
	switch name {
	case "launcher":
		return ui.TabLauncher
	case "monitor":
		return ui.TabMonitor
	case "models":
		return ui.TabModels
	default:
		return ui.TabProfiles
	}
}
