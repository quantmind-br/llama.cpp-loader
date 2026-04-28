// Package pages holds tab page implementations.
package pages

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
)

// profileDraft is the editor's mutable state, mapped from huh form back to a Profile on save.
type profileDraft struct {
	ID          string // immutable once created
	Name        string
	Description string
	Model       string
	NGL         string
	CtxSize     string
	Port        string
	FlashAttn   bool
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

func buildEditorForm(d *profileDraft) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&d.Name),
			huh.NewInput().Title("Description").Value(&d.Description),
			huh.NewInput().Title("Model path").Value(&d.Model),
		),
		huh.NewGroup(
			huh.NewInput().Title("ngl (gpu layers)").Value(&d.NGL),
			huh.NewInput().Title("ctx-size").Value(&d.CtxSize),
			huh.NewInput().Title("port").Value(&d.Port),
			huh.NewConfirm().Title("flash-attn?").Value(&d.FlashAttn).Affirmative("Yes").Negative("No"),
		),
	).WithShowHelp(true)
}
