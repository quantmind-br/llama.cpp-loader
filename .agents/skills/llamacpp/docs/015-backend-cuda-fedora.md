---
title: CUDA FEDORA
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/CUDA-FEDORA.md
source: git
fetched_at: 2026-04-28T09:49:08.915771312-03:00
rendered_js: false
word_count: 451
summary: This document provides a step-by-step guide for installing and configuring the Nvidia CUDA toolkit within a Fedora toolbox container to avoid system conflicts.
tags:
    - cuda
    - fedora
    - toolbox
    - nvidia
    - containerization
    - development-environment
category: tutorial
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Setting Up CUDA on Fedora

Install [Nvidia CUDA](https://docs.nvidia.com/cuda/) in a toolbox container. Applies to:

- [Fedora Workstation](https://fedoraproject.org/workstation/)
- [Atomic Desktops for Fedora](https://fedoraproject.org/atomic-desktops/)
- [Fedora Spins](https://fedoraproject.org/spins)
- [Other distributions](https://containertoolbx.org/distros/) (RHEL >= 8.5, Arch, Ubuntu)

## Prerequisites

- **Toolbox installed on host** — Fedora Silverblue/Workstation include it by default; others need the [toolbox package](https://containertoolbx.org/install/).
- **NVIDIA drivers and GPU on host** (recommended) — Fedora hosts can use the [RPM Fusion Repository](https://rpmfusion.org/Howto/NVIDIA).
- **Internet connectivity**

> [!info] Latest release: Fedora 41. [Fedora 41 CUDA Repository](https://developer.download.nvidia.com/compute/cuda/repos/fedora41/x86_64/).

## Create a Fedora Toolbox Environment

> [!note] Toolbox is available for other systems. Podman or Docker can also work.

1. Create a Fedora 41 Toolbox:

   ```bash
   toolbox create --image registry.fedoraproject.org/fedora-toolbox:41 --container fedora-toolbox-41-cuda
   ```

2. Enter the Toolbox:

   ```bash
   toolbox enter --container fedora-toolbox-41-cuda
   ```

Inside you have root privileges; packages won't affect the host.

## Install Essential Development Tools

```bash
sudo dnf distro-sync
sudo dnf install vim-default-editor --allowerasing   # optional
sudo dnf install @c-development @development-tools cmake
```

The `--allowerasing` flag removes the conflicting `nano-default-editor` package.

## Add the CUDA Repository

```bash
sudo dnf config-manager addrepo --from-repofile=https://developer.download.nvidia.com/compute/cuda/repos/fedora41/x86_64/cuda-fedora41.repo
sudo dnf distro-sync
```

## Install Nvidia Driver Libraries

Check if the host supplies driver libraries into the toolbox:

```bash
ls -la /usr/lib64/libcuda.so.1
```

### If `libcuda.so.1` is missing — install on guest:

```bash
sudo dnf install nvidia-driver-cuda nvidia-driver-libs nvidia-driver-cuda-libs nvidia-persistenced
```

### If `libcuda.so.1` exists — update the guest RPM database:

The host already supplies the drivers. Update the DB so the guest recognizes them:

1. Download the nvidia packages (with dependencies):

   ```bash
   sudo dnf download --destdir=/tmp/nvidia-driver-libs --resolve --arch x86_64 nvidia-driver-cuda nvidia-driver-libs nvidia-driver-cuda-libs nvidia-persistenced
   ```

2. Update the RPM database only (`--justdb`):

   ```bash
   sudo rpm --install --verbose --hash --justdb /tmp/nvidia-driver-libs/*
   ```

3. Verify — the following should report "already installed":

   ```bash
   sudo dnf install nvidia-driver-cuda nvidia-driver-libs nvidia-driver-cuda-libs nvidia-persistenced
   ```

   ```
   Package "nvidia-driver-cuda-3:570.124.06-1.fc41.x86_64" is already installed.
   Package "nvidia-driver-libs-3:570.124.06-1.fc41.x86_64" is already installed.
   Package "nvidia-driver-cuda-libs-3:570.124.06-1.fc41.x86_64" is already installed.
   Package "nvidia-persistenced-3:570.124.06-1.fc41.x86_64" is already installed.

   Nothing to do.
   ```

## Install the CUDA Meta-Package

```bash
sudo dnf install cuda
```

## Configure the Environment

Add CUDA binaries to `PATH`:

1. Create a profile script:

   ```bash
   sudo sh -c 'echo "export PATH=\$PATH:/usr/local/cuda/bin" >> /etc/profile.d/cuda.sh'
   ```

   The `/etc/profile.d/` directory is container-specific (not shared with host). The `\` before `$PATH` ensures correct variable expansion.

2. Make it executable and source it:

   ```bash
   sudo chmod +x /etc/profile.d/cuda.sh
   source /etc/profile.d/cuda.sh
   ```

   The script ensures CUDA binaries are in `PATH` for all future sessions.

## Verify the Installation

```bash
nvcc --version
```

Expected output:

```
nvcc: NVIDIA (R) Cuda compiler driver
Copyright (c) 2005-2025 NVIDIA Corporation
Built on Fri_Feb_21_20:23:50_PST_2025
Cuda compilation tools, release 12.8, V12.8.93
Build cuda_12.8.r12.8/compiler.35583870_0
```

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Installation failures (conflicting files/missing deps) | Read error messages carefully. Use `rpm --excludepath` for manual RPM installs. |
| NVIDIA driver host passthrough bug (missing shared lib) | Restart the container: `podman container restart --all` |
| `nvcc` not found after install | Verify `/usr/local/cuda/bin` is in `PATH` (`echo $PATH`). Re-source the profile script or open a new terminal. |

## Additional Notes

- **Updating CUDA**: Monitor official NVIDIA repos for Fedora version updates and adjust `dnf` config accordingly.
- **Building llama.cpp**: With CUDA installed, follow the [[027-build|build instructions]] to compile with CUDA support. Ensure CUDA-specific build flags are set correctly.
- **Toolbox isolation**: System files and configs inside the toolbox are separate from the host. The home directory is shared by default.

> [!warning] Manually installing/modifying system packages can destabilize the container. Back up important data before major changes, especially since your home folder is shared with the toolbox.

## References

- [Fedora Toolbox Documentation](https://docs.fedoraproject.org/en-US/fedora-silverblue/toolbox/)
- [NVIDIA CUDA Installation Guide](https://docs.nvidia.com/cuda/cuda-installation-guide-linux/index.html)
- [Podman Documentation](https://podman.io/get-started)

#cuda #fedora #toolbox #nvidia #containerization
