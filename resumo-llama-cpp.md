# Resumo: Setup llama.cpp para Inferência Local

## Hardware Alvo
- **GPU:** RTX 3090 (24GB VRAM)
- **OS:** Arch Linux

## llama.cpp

### Versão
- Repo oficial: `ggml-org/llama.cpp` (renomeado de `ggerganov/llama.cpp`)
- Sem LTS — usar latest release ou git master

### Instalação Escolhida (NVIDIA/CUDA)

**Pré-requisitos:**
```bash
sudo pacman -S nvidia nvidia-utils cuda cuda-tools
nvidia-smi      # verifica driver
nvcc --version  # verifica CUDA
```

**Pacote AUR:**
```bash
yay -S llama.cpp-cuda
```

### Outros Backends (referência)
| GPU | Pacote AUR |
|-----|------------|
| CPU only | `llama.cpp` |
| AMD ROCm | `llama.cpp-hip` |
| AMD/Intel/Vulkan | `llama.cpp-vulkan` |
| Intel SYCL | `llama.cpp-sycl-fp16` |

## Modelos GGUF

### Fonte Recomendada
- HuggingFace: usuário `bartowski` (padrão ouro)

### Sugestões para 24GB VRAM
| Modelo | Quant | VRAM |
|--------|-------|------|
| Llama 3.3 70B | Q4_K_M | ~22GB |
| Qwen 2.5 32B | Q5_K_M | ~22GB |
| Qwen 2.5 Coder 32B | Q4_K_M | ~18GB |
| Mistral Small 24B | Q5_K_M | ~17GB |
| Qwen 2.5 14B | Q6_K | ~12GB |

## LM Studio (compartilhamento de modelos)

- **Path Linux:** `~/.lmstudio/models/<publisher>/<repo>/<file>.gguf`
- Pode apontar `llama-server -m` direto pro arquivo do LM Studio (sem duplicar)
- Settings → Models Directory = muda local

## Configuração de Parâmetros

### 3 Métodos
1. **CLI flags** — direto na linha de comando
2. **Config JSON** — `llama-server --config config.json`
3. **Env vars** — `LLAMA_ARG_<FLAG_UPPER>`

### Params de Load (boot)
| Flag | Função | Default |
|------|--------|---------|
| `-m` | Path do GGUF | obrigatório |
| `-ngl N` | Camadas na GPU (99 = tudo) | 0 |
| `-c N` | Context size | 4096 |
| `-b N` | Logical batch | 2048 |
| `-ub N` | Physical batch | 512 |
| `--flash-attn` | Flash attention | off |
| `-t N` | Threads CPU | auto |
| `-np N` | Slots paralelos | 1 |
| `--mlock` | Trava em RAM | off |
| `-ctk` / `-ctv` | KV cache quant (`f16`/`q8_0`/`q4_0`) | f16 |
| `-sm` | Split GPUs (`layer`/`row`/`none`) | layer |
| `-ts a,b` | Tensor split multi-GPU | uniforme |

### Params de Sample (runtime, via HTTP)
- `temperature`, `top_k`, `top_p`, `min_p`, `repeat_penalty`, `n_predict`

### Stack Recomendada para RTX 3090
```bash
llama-server \
  -m ~/models/Qwen2.5-32B-Instruct-Q5_K_M.gguf \
  -ngl 99 \
  -c 16384 \
  -b 2048 -ub 512 \
  --flash-attn \
  -ctk q8_0 -ctv q8_0 \
  -np 2 \
  --port 8080
```

### Otimizações VRAM
- `--flash-attn` → sempre ligar (grátis)
- `-ctk q8_0 -ctv q8_0` → ~50% economia no KV cache
- Context grande demais → reduz `-ngl` ou quant menor

## Endpoints

- **Web UI:** `http://localhost:8080`
- **API OpenAI-compatible:** `http://localhost:8080/v1/chat/completions`
- Adicionar `--api-key sk-local` p/ proteger

## Comandos Úteis

```bash
llama-server --help | less   # lista todas flags (fonte verdade)
lspci | grep -E 'VGA|3D'     # identifica GPU
```
