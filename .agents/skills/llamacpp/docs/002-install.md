---
title: Install pre-built version of llama.cpp
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/install.md
source: git
fetched_at: 2026-04-28T09:49:29.614061938-03:00
rendered_js: false
word_count: 79
summary: This document provides instructions for installing pre-built versions of llama.cpp across various operating systems using different package managers.
tags:
    - llama-cpp
    - installation-guide
    - package-management
    - winget
    - homebrew
    - macports
    - nix
category: guide
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
# Install pre-built version of llama.cpp

| Install via | Windows | Mac | Linux |
|-------------|---------|-----|-------|
| Winget      | ✅      |     |       |
| Homebrew    |         | ✅  | ✅    |
| MacPorts    |         | ✅  |       |
| Nix         |         | ✅  | ✅    |

## Winget (Windows)

```sh
winget install llama.cpp
```

Auto-updated with new releases. [Info](https://github.com/ggml-org/llama.cpp/issues/8188)

## Homebrew (Mac and Linux)

```sh
brew install llama.cpp
```

Auto-updated with new releases. [Info](https://github.com/ggml-org/llama.cpp/discussions/7668)

## MacPorts (Mac)

```sh
sudo port install llama.cpp
```

[Details](https://ports.macports.org/port/llama.cpp/details/)

## Nix (Mac and Linux)

Flake-enabled:

```sh
nix profile install nixpkgs#llama-cpp
```

Non-flake:

```sh
nix-env --file '<nixpkgs>' --install --attr llama-cpp
```

Auto-updated within [nixpkgs](https://github.com/NixOS/nixpkgs/blob/nixos-24.05/pkgs/by-name/ll/llama-cpp/package.nix#L164).
