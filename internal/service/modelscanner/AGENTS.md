# PACKAGE: modelscanner

## OVERVIEW
Filesystem scanner for GGUF models. Recursively walks configured paths, extracts quant/params from filenames, and reads parameter counts from GGUF headers via a lightweight binary parser.

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Scanner interface | `modelscanner.go` | Single method: `Scan(ctx, paths) -> channel` |
| Walk + event emission | `scanner.go` | Buffered channel (64); emits File/Progress/Error/Done events |
| GGUF binary parsing | `gguf.go` | Reads magic, header, metadata KV pairs. Scalar types only. |
| Quant/param extraction | `quant.go` | Regex from filename. Quant order matters (specific first). |
| Tests | `*_test.go` | `writeGGUFFile` helper builds synthetic GGUFs in `t.TempDir()` |

## CONVENTIONS
- Scanner returns a buffered channel. Caller **must drain until close**.
- Event types: `ScanEventFile`, `ScanEventProgress`, `ScanEventError`, `ScanEventDone`.
- `buildModelFile` tries GGUF header first, falls back to filename parsing for params.
- All GGUF errors are swallowed (return `""`), not propagated to the UI.

## ANTI-PATTERNS
- **Never** leave a Scanner channel undrained. The background goroutine will leak.
- Do not add array type support to `skipGGUFValue` without understanding GGUF array encoding.
- Do not reorder `quantRe` alternatives carelessly. `"Q4_K_M"` must match before `"Q4"`.

## NOTES
- Metadata scan capped at 128 KV pairs. `general.parameter_count` and `general.size_label` are usually near the top.
- String reads capped at 1 MiB to defend against corrupt headers.
- `formatParams` rounds: `>=1B` uses B, `>=1M` uses M, else raw integer.
