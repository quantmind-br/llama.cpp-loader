# Documentation Gaps Analysis Report

**Project:** llama-cpp-loader
**Date:** 2026-04-30
**Analyzer:** Claude Code (ideation-documentation skill)

---

## Executive Summary

O projeto `llama-cpp-loader` possui **documentação de desenvolvimento razoável** (AGENTS.md, design specs em `docs/superpowers/`), mas carece de documentação voltada para **usuários finais e novos contribuidores**. Não existe README.md, nenhum guia de instalação, e a documentação de configuração é inexistente. A cobertura de comentários Go é ~87% em funções exportadas (bom), mas há lacunas em código complexo interno e em alguns métodos de UI pages.

| Métrica | Valor |
|---------|-------|
| Arquivos Go analisados | 43 (sem testes) |
| Funções exportadas | ~243 |
| Comentários de função exportada | ~212 |
| Cobertura de funções exportadas | ~87% |
| Package comments | 17 |
| Tipos exportados | ~66 |
| README.md | **AUSENTE** |

---

## Current Documentation Assessment

### README.md
- **Status:** Missing
- **Quality:** N/A
- **Key Gaps:** O projeto inteiramente carece de um README. Quem chega ao repositório não tem visão do que o projeto faz, como instalar, nem como usar.

### API Documentation (Go doc)
- **Funções documentadas:** ~212 / ~243 (~87%)
- **Package comments:** 17/17 packages possuem comentário de package (excelente)
- **Tipos documentados:** Boa cobertura em `domain/`, `config/`, `monitor/`, `processmgr/`
- **Lacunas:** Funções privadas complexas em `internal/ui/pages/` e `internal/service/processmgr/` carecem de explicação

### Inline Comments
- **Densidade geral:** Média-Alta
- **Pontos fortes:**
  - `internal/service/modelscanner/gguf.go`: Comentários excelentes sobre parsing binário GGUF
  - `internal/service/processmgr/processmgr.go`: Interfaces e sentinel errors bem documentadas
  - `internal/domain/`: Todos os tipos comentados
- **Pontos fracos:**
  - `internal/ui/pages/profiles_editor.go`: `buildEditorForm()` sem comentário de propósito
  - `internal/ui/pages/profiles.go`: Lógica de `Update()` complexa (160+ linhas) sem comentários de seção
  - `internal/service/monitor/subscribe.go`: Goroutine orchestration sem explicação de lifecycle

### Examples & Tutorials
- **Status:** Missing
- Nenhum guia "Getting Started", nenhum exemplo de config.toml, nenhum screenshot ou GIF da TUI em ação.

### Architecture Documentation
- **Status:** Partial
- Existe `AGENTS.md` (project knowledge base) e `docs/superpowers/specs/` com design specs — ambos voltados para agentes de IA durante desenvolvimento.
- Falta documentação arquitetural para contribuidores humanos (diagrama de camadas, fluxo de dados, decisões de design).

### Troubleshooting
- **Status:** Missing (para usuários)
- Existe `BUG_REPORT.md` com bugs conhecidos — mas é um artefato interno de desenvolvimento, não um guia para usuários.

---

## Documentation Gaps

### High Priority

#### DOC-001: README.md Ausente

**Category:** readme
**Target Audience:** users, contributors

**Affected Areas:**
- `/README.md` (não existe)

**Current State:**
O repositório não possui README.md. A página inicial do projeto no GitHub estará vazia.

**Proposed Content:**
- Descrição do projeto: "TUI para gerenciamento de profiles e processos llama-server"
- Screenshot/GIF da interface
- Requisitos (Go 1.26+, llama-server no PATH)
- Instalação: `make build` ou `go install`
- Uso básico: navegação por tabs, criar profile, fazer launch
- Estrutura de diretórios de config (`~/.config/llama-cpp-loader/`)
- Atalhos globais (`1-4`, `Tab`, `q`, `?`)

**Rationale:**
README é o primeiro contato de qualquer pessoa com o projeto. Sem ele, não há adoção nem contribuições.

**Estimated Effort:** small

---

#### DOC-002: Documentação de Configuração (config.toml)

**Category:** examples
**Target Audience:** users

**Affected Areas:**
- `internal/config/config.go`
- `~/.config/llama-cpp-loader/config.toml` (runtime)

**Current State:**
O código define defaults em `applyDefaults()`, mas não há documentação do formato TOML nem explicação de cada campo.

**Proposed Content:**
Arquivo `docs/config.md` ou seção no README com:
```toml
[paths]
profiles_dir = "~/.config/llama-cpp-loader/profiles"
log_dir = "~/.local/state/llama-cpp-loader/logs"
state_dir = "~/.local/state/llama-cpp-loader"

[models]
search_paths = ["~/.lmstudio/models", "~/models"]

[ui]
default_tab = "profiles"
keybindings = "default"
```

**Rationale:**
Usuários precisam saber como customizar paths de modelos e profiles. Sem docs, só lendo o código-fonte.

**Estimated Effort:** trivial

---

#### DOC-003: Guia de Instalação e Primeiros Passos

**Category:** examples
**Target Audience:** users

**Affected Areas:**
- README (novo)
- Potencial `docs/getting-started.md`

**Current State:**
Makefile existe (`make build`, `make install`, `make tests`), mas não há instruções passo a passo.

**Proposed Content:**
1. Pré-requisitos: Go 1.26+, llama-server compilado e no PATH
2. Clone e build
3. Primeira execução (config.toml auto-gerado)
4. Criar primeiro profile (tab Profiles → `n`)
5. Fazer um launch (tab Launcher → Enter)
6. Monitorar (tab Monitor)

**Rationale:**
Onboarding blocker. Um usuário novo não consegue usar a aplicação sem adivinhar o workflow.

**Estimated Effort:** small

---

#### DOC-004: Comentários em Código Complexo de UI Pages

**Category:** inline_comments
**Target Audience:** developers, contributors

**Affected Areas:**
- `internal/ui/pages/profiles.go` — método `Update()` (~200 linhas, múltiplos estados: list, edit, confirm, picker, advanced)
- `internal/ui/pages/monitor.go` — goroutine lifecycle, subscription management
- `internal/ui/pages/launcher.go` — estado do spinner + wait healthy

**Current State:**
- `profiles.go:Update()` não tem comentários de seção para os grandes blocos de lógica (list selection, edit mode, picker overlay, confirm dialogs, advanced tab)
- Estados transitórios (editing, pickerActive, confirmDelete, filterMode) não documentados

**Proposed Content:**
Comentários de seção dentro de `Update()`:
```go
// Phase 1: Handle page-global messages (window resize, flash clear)
// Phase 2: Handle picker overlay (model selection)
// Phase 3: Handle form editing (huh form state machine)
// Phase 4: Handle list navigation and actions
// Phase 5: Forward non-key messages to active sub-model
```

**Rationale:**
Contribuidores novos precisam entender a máquina de estados de cada page para não introduzir regressões no input routing.

**Estimated Effort:** small

---

### Medium Priority

#### DOC-005: Documentação da Arquitetura para Contribuidores

**Category:** architecture
**Target Audience:** contributors, maintainers

**Affected Areas:**
- Nova pasta `docs/architecture/` ou seção em README

**Current State:**
- `AGENTS.md` tem um overview estrutural excelente, mas é escrito para agentes de IA
- `docs/superpowers/specs/` tem o design spec original, mas está em português e focado em decisões de produto
- Não há diagrama de fluxo de dados nem explicação de como as camadas se comunicam

**Proposed Content:**
- Diagrama de camadas (cmd → ui → service → domain)
- Fluxo de mensagens bubbletea (root → page → component)
- Explicação do process lifecycle (foreground vs background, recovery de instances.json)
- Explicação do input routing (global shortcuts vs page capture)

**Rationale:**
A arquitetura é sofisticada (TUI + process management + monitoring). Documentação arquitetural reduz o tempo para contribuidores entenderem onde fazer mudanças.

**Estimated Effort:** medium

---

#### DOC-006: Documentação de Interfaces Exportadas

**Category:** api_docs
**Target Audience:** developers, contributors

**Affected Areas:**
- `internal/service/processmgr/manager.go` — `fsManager` (struct privada) não tem comentários de campo
- `internal/service/profilestore/fs_store.go` — `fsStore` não documentada
- `internal/service/monitor/subscribe.go` — funções `start*Poller`, `run*Pump` sem comentários

**Current State:**
As interfaces públicas (`Manager`, `Store`, `Scanner`, `Validator`) estão bem documentadas. Mas as implementações concretas (`fsManager`, `fsStore`) carecem de comentários em métodos privados e em funções helper.

**Proposed Content:**
```go
// fsManager implements processmgr.Manager using os/exec.
// It tracks running instances in memory and persists background
// instances to instances.json for recovery across TUI restarts.
type fsManager struct { ... }
```

**Rationale:**
Contribuidores que precisam debuggar process lifecycle ou monitor subscription precisam entender a implementação, não só a interface.

**Estimated Effort:** small

---

#### DOC-007: Comentários no Parser de GGUF

**Category:** inline_comments
**Target Audience:** developers, contributors

**Affected Areas:**
- `internal/service/modelscanner/gguf.go`

**Current State:**
O arquivo já tem bons comentários (magic number, header, type IDs). Mas `skipGGUFValue()` e `formatParams()` carecem de explicação.

**Proposed Content:**
```go
// skipGGUFValue advances the reader past a metadata value of the given type.
// Returns false if the type is unknown (caller should abort the scan).
func skipGGUFValue(r io.Reader, typeID uint32) bool { ... }
```

**Rationale:**
Parsing binário é inerentemente complexo e propenso a erros silenciosos. Comentários claros evitam regressões.

**Estimated Effort:** trivial

---

#### DOC-008: Documentação de Atalhos de Teclado

**Category:** examples
**Target Audience:** users

**Affected Areas:**
- `internal/ui/components/help.go` (conteúdo do help modal)
- README

**Current State:**
O modal de ajuda (`?`) existe e lista atalhos. Mas não há documentação externa que um usuário possa consultar sem abrir a aplicação.

**Proposed Content:**
Tabela em README ou `docs/shortcuts.md`:
| Contexto | Tecla | Ação |
|----------|-------|------|
| Global | `1-4` | Trocar tab |
| Global | `Tab` | Próxima tab |
| Global | `q` | Sair |
| Global | `?` | Ajuda |
| Profiles | `n` | Novo profile |
| Profiles | `enter` | Editar profile |
| Profiles | `s` | Salvar |
| ... | ... | ... |

**Rationale:**
Usuários precisam de uma referência rápida. O modal de ajuda é útil mas limitado em espaço.

**Estimated Effort:** trivial

---

### Low Priority

#### DOC-009: Guia de Contribuição

**Category:** readme
**Target Audience:** contributors

**Affected Areas:**
- `CONTRIBUTING.md` (novo)

**Current State:**
Não existe.

**Proposed Content:**
- Como buildar: `make build`
- Como rodar testes: `make tests`
- Como atualizar golden tests: `go test ./... -update`
- Estrutura de branches e commits
- Regras de input routing (anti-pattern do projeto)

**Rationale:**
Importante para projeto open-source, mas não é blocker para uso.

**Estimated Effort:** small

---

#### DOC-010: Documentação de Troubleshooting

**Category:** troubleshooting
**Target Audience:** users, maintainers

**Affected Areas:**
- `docs/troubleshooting.md` (novo)

**Current State:**
`BUG_REPORT.md` lista bugs conhecidos, mas não é um guia de troubleshooting.

**Proposed Content:**
- "llama-server não encontrado" → verificar PATH
- "Porta em uso" → como alterar porta no profile
- "Modelo não aparece no browser" → verificar `search_paths` no config.toml
- "Instância background não recupera" → verificar `instances.json`
- "Erro de validação de flag" → explicação do schema version-aware

**Rationale:**
Reduz issues abertos por usuários e tempo de suporte.

**Estimated Effort:** small

---

## Documentation Coverage Summary

| Category | Status | Cobertura |
|----------|--------|-----------|
| README | Missing | 0% |
| API Docs | Needs Improvement | ~87% (funções exportadas) |
| Inline Comments | Needs Improvement | Alta em domain/service, baixa em pages |
| Examples | Missing | 0% |
| Architecture | Partial | Existe para agentes, não para humanos |
| Troubleshooting | Missing | 0% |

---

## Statistics

| Métrica | Valor |
|---------|-------|
| Total de arquivos Go analisados | 43 |
| Funções exportadas documentadas | ~212 |
| Funções exportadas sem documentação | ~31 |
| Cobertura de documentação de API | ~87% |
| Package comments | 17/17 (100%) |
| Arquivos de documentação existentes | 14 (.md) |
| Arquivos de documentação voltados a usuários | 0 |

| Prioridade | Count |
|------------|-------|
| High | 4 |
| Medium | 4 |
| Low | 2 |

| Categoria | Count |
|-----------|-------|
| README | 2 (DOC-001, DOC-009) |
| API Docs | 1 (DOC-006) |
| Inline Comments | 2 (DOC-004, DOC-007) |
| Examples | 2 (DOC-002, DOC-003, DOC-008) |
| Architecture | 1 (DOC-005) |
| Troubleshooting | 1 (DOC-010) |
