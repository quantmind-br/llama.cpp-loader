// Package pages holds tab page implementations.
package pages

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

type subTab int

const (
	subTabEssentials subTab = iota
	subTabAdvanced
)

func (s subTab) String() string {
	if s == subTabEssentials {
		return "Essentials"
	}
	return "Advanced"
}

// profileDraft is the editor's mutable state, mapped from huh form back to a Profile on save.
type profileDraft struct {
	ID          string // immutable once created
	Name        string
	Description string
	Model       string
	NGL         string
	CtxSize     string
	BatchSize   string
	UBatchSize  string
	Port        string
	FlashAttn   bool
	CacheTypeK  string
	CacheTypeV  string
	isNew       bool
}

func argString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func argBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func buildEditorForm(d *profileDraft, schema domain.FlagSchema) *huh.Form {
	cacheOpts := selectOptions(schema, "cache-type-k", []string{"f16", "q8_0", "q4_0"})
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&d.Name),
			huh.NewInput().Title("Description").Value(&d.Description),
			huh.NewInput().Title("Model path (.gguf)").Value(&d.Model),
		),
		huh.NewGroup(
			huh.NewInput().Title(labelWithHelp(schema, "n-gpu-layers", "ngl (gpu layers)")).Value(&d.NGL),
			huh.NewInput().Title(labelWithHelp(schema, "ctx-size", "ctx-size")).Value(&d.CtxSize),
			huh.NewInput().Title(labelWithHelp(schema, "batch-size", "batch-size")).Value(&d.BatchSize),
			huh.NewInput().Title(labelWithHelp(schema, "ubatch-size", "ubatch-size")).Value(&d.UBatchSize),
			huh.NewInput().Title(labelWithHelp(schema, "port", "port")).Value(&d.Port),
			huh.NewConfirm().Title(labelWithHelp(schema, "flash-attn", "flash-attn?")).Value(&d.FlashAttn).Affirmative("Yes").Negative("No"),
			huh.NewSelect[string]().Title("cache-type-k").Options(toOptions(cacheOpts)...).Value(&d.CacheTypeK),
			huh.NewSelect[string]().Title("cache-type-v").Options(toOptions(cacheOpts)...).Value(&d.CacheTypeV),
		),
	).WithShowHelp(true)
}

func labelWithHelp(schema domain.FlagSchema, name, fallback string) string {
	if spec, ok := schema.Lookup(name); ok && spec.HelpText != "" {
		if spec.Default != nil {
			return fmt.Sprintf("%s — %s (default %v)", fallback, spec.HelpText, spec.Default)
		}
		return fmt.Sprintf("%s — %s", fallback, spec.HelpText)
	}
	return fallback
}

func selectOptions(schema domain.FlagSchema, name string, fallback []string) []string {
	if spec, ok := schema.Lookup(name); ok && len(spec.EnumValues) > 0 {
		return spec.EnumValues
	}
	return fallback
}

func toOptions(values []string) []huh.Option[string] {
	out := make([]huh.Option[string], 0, len(values))
	for _, v := range values {
		out = append(out, huh.NewOption(v, v))
	}
	return out
}

func newAdvancedTable(schema domain.FlagSchema, width, height int) table.Model {
	helpWidth := width - 24 - 8 - 12 - 8
	if helpWidth < 16 {
		helpWidth = 16
	}
	cols := []table.Column{
		{Title: "Flag", Width: 24},
		{Title: "Type", Width: 8},
		{Title: "Default", Width: 12},
		{Title: "Help", Width: helpWidth},
	}
	rows := schemaRows(schema)
	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithFocused(true), table.WithHeight(height))
	return t
}

func schemaRows(schema domain.FlagSchema) []table.Row {
	names := make([]string, 0, len(schema.Flags))
	for k := range schema.Flags {
		names = append(names, k)
	}
	sort.Strings(names)
	rows := make([]table.Row, 0, len(names))
	for _, name := range names {
		spec := schema.Flags[name]
		rows = append(rows, table.Row{
			spec.Long,
			typeLabel(spec.Type),
			fmt.Sprintf("%v", spec.Default),
			truncate(spec.HelpText, 80),
		})
	}
	return rows
}

func typeLabel(t domain.FlagType) string {
	switch t {
	case domain.FlagTypeBool:
		return "bool"
	case domain.FlagTypeInt:
		return "int"
	case domain.FlagTypeFloat:
		return "float"
	case domain.FlagTypeEnum:
		return "enum"
	default:
		return "string"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// filterRows returns rows whose Flag column contains q (case-insensitive).
func filterRows(all []table.Row, q string) []table.Row {
	if q == "" {
		return all
	}
	q = strings.ToLower(q)
	out := make([]table.Row, 0, len(all))
	for _, r := range all {
		if strings.Contains(strings.ToLower(r[0]), q) {
			out = append(out, r)
		}
	}
	return out
}
