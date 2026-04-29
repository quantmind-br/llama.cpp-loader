// Command llama-cpp-loader launches the TUI for managing llama.cpp profiles.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/config"
	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/llamahelp"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/modelscanner"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/validator"
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

	schema, schemaWarn := loadSchema()
	scanner := modelscanner.New()

	mgr := processmgr.New(processmgr.Config{
		LogDir:       cfg.Paths.LogDir,
		RegistryPath: filepath.Join(cfg.Paths.StateDir, "instances.json"),
		LastUsedSink: store,
	})
	if err := mgr.Reconcile(); err != nil {
		fmt.Fprintf(os.Stderr, "instance recovery: %v\n", err)
	}

	val := validator.New()

	profilesPage := pages.NewProfilesPage(store, schema).
		WithModelScanner(scanner, cfg.Models.SearchPaths)
	modelsPage := pages.NewModelsPage(scanner, cfg.Models.SearchPaths)
	launcherPage := pages.NewLauncherPage(store, mgr, val).SetSchema(schema)

	root := ui.NewRoot(parseTab(cfg.UI.DefaultTab)).
		WithProfilesPage(profilesPage).
		WithModelsPage(modelsPage).
		WithLauncherPage(launcherPage)
	if schemaWarn != "" {
		root = root.WithStatusWarn(schemaWarn)
	}

	// Background llama-server processes intentionally survive TUI exit;
	// processmgr.Reconcile restores them at next boot from instances.json.
	prog := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}

// loadSchema attempts to parse llama-server --help. On failure (binary absent,
// timeout, parse error) it returns the embedded fallback and a warning string
// suitable for the status bar.
func loadSchema() (domain.FlagSchema, string) {
	parser := llamahelp.NewExecParser()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	schema, err := parser.Parse(ctx)
	if err != nil {
		return llamahelp.EmbeddedSchema(), fmt.Sprintf("schema fallback: %v", err)
	}
	return schema, ""
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
