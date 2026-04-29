package components

import "github.com/charmbracelet/glamour"

// HelpMarkdown é o conteúdo da modal de help acessível via `?` em qualquer
// página. Atualizado quando keybindings mudam.
const HelpMarkdown = `# llama-cpp-loader — Keybindings

## Global

- ` + "`1`" + `–` + "`4`" + ` — switch directly to a tab
- ` + "`Tab`" + ` — next tab     ` + "`Shift+Tab`" + ` — previous tab
- ` + "`?`" + ` — toggle this help
- ` + "`q`" + ` / ` + "`Ctrl+C`" + ` — quit (background instances survive)

## Profiles tab

- ` + "`n`" + ` — new profile     ` + "`d`" + ` — duplicate
- ` + "`x`" + ` — delete         ` + "`s`" + ` — save
- ` + "`L`" + ` — launch directly from selected profile
- ` + "`/`" + ` — filter

## Launcher tab

- ` + "`b`" + ` — toggle background/foreground (default background)
- ` + "`enter`" + ` — launch selected profile
- ` + "`k`" + ` — kill the most recent launched instance
- ` + "`r`" + ` — refresh profile list

## Monitor tab

- ` + "`Tab`" + ` — cycle Logs / Slots / Metrics sub-views
- ` + "`Space`" + ` — pause/resume log scroll
- ` + "`k`" + ` — kill selected instance
- ` + "`r`" + ` — restart selected instance (Kill + Launch)

## Models tab

- ` + "`R`" + ` — rescan all configured paths
- ` + "`/`" + ` — filter
- ` + "`enter`" + ` — actions: use in new profile / existing profile / reveal path
`

// RenderHelp retorna o markdown HelpMarkdown renderizado via glamour.
// width informa ao renderer o tamanho da viewport em colunas (afeta wrap).
func RenderHelp(width int) (string, error) {
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}
	return r.Render(HelpMarkdown)
}
