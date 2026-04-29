---
title: Backend Snapdragon Windows
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/snapdragon/windows.md
source: git
fetched_at: 2026-04-28T09:49:18.194640091-03:00
rendered_js: false
word_count: 309
summary: This document provides instructions for setting up the development environment on Snapdragon Windows devices, including driver installation, SDK configuration, and the process for test-signing HTP ops libraries.
tags:
    - snapdragon
    - windows-driver
    - npu-setup
    - hexagon-sdk
    - adreno-sdk
    - test-signing
    - driver-installation
category: guide
optimized: true
optimized_at: '2026-04-28T00:00:00Z'
---
# Backend Snapdragon Windows

## Overview

To use Hexagon NPU on Snapdragon Windows devices, the HTP Ops libraries (e.g. `libggml-htp-v73.so`) must be included in a `.cat` file digitally signed with a trusted certificate. This guide covers installing GPU/NPU drivers and SDKs, generating personal certificate files (`.pfx`), and configuring test-signing.

## Install the latest Adreno OpenCL SDK

Use the trimmed CI version:

    https://github.com/snapdragon-toolchain/opencl-sdk/releases/download/v2.3.2/adreno-opencl-sdk-v2.3.2-arm64-wos.tar.xz

Or download the complete version from:

    https://softwarecenter.qualcomm.com/catalog/item/Adreno_OpenCL_SDK?version=2.3.2

Extract into:

```
c:\Qualcomm\OpenCL_SDK\2.3.2
```

## Install the latest Hexagon SDK Community Edition

Use the trimmed CI version:

    https://github.com/snapdragon-toolchain/hexagon-sdk/releases/download/v6.4.0.2/hexagon-sdk-v6.4.0.2-arm64-wos.tar.xz

Or download the complete version from:

    https://softwarecenter.qualcomm.com/catalog/item/Hexagon_SDK?version=6.4.0.2

Extract into:

```
c:\Qualcomm\Hexagon_SDK\6.4.0.2
```

## Install the latest Adreno GPU driver

Download from:

    https://softwarecenter.qualcomm.com/catalog/item/Windows_Graphics_Driver

After installation and reboot, verify the GPU appears in `Device Manager` under `Display Adapters`.

## Install the latest Qualcomm NPU driver

Download from:

    https://softwarecenter.qualcomm.com/catalog/item/Qualcomm_HND

After installation and reboot, verify the Hexagon NPU appears in `Device Manager` under `Neural Processors`.

If the device is missing, install all components (`qcnspmcdm8380`, `qcnspmcdm8380_ext`) manually. Components extract to:

```
c:\QCDrivers\qcnspmcdm...
```

## Enable NPU driver test signatures

> [!warning]
> The following steps are required **only** for the Hexagon NPU. The Adreno GPU backend does not require test signatures.

### Enable testsigning

```
> bcdedit /set TESTSIGNING ON
```

> [!note]
> Secure Boot may need to be disabled for this to work.

Verify after reboot:

```
> bcdedit /enum
...
testsigning             Yes
...
```

Microsoft reference: https://learn.microsoft.com/en-us/windows-hardware/drivers/install/the-testsigning-boot-configuration-option

### Create personal certificate

Tools are part of Windows SDK / Driver Kit (installed with Visual Studio), typically at:

```
c:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0
```

Create a self-signed certificate:

```
> cd c:\Users\MyUser
> mkdir Certs
> cd Certs
> makecert -r -pe -ss PrivateCertStore -n CN=GGML.HTP.v1 -eku 1.3.6.1.5.5.7.3.3 -sv ggml-htp-v1.pvk ggml-htp-v1.cer
> pvk2pfx.exe -pvk ggml-htp-v1.pvk -spc ggml-htp-v1.cer -pfx ggml-htp-v1.pfx
```

Replace `MyUser` with your username. Import the PFX into `Trusted Root Certification Authorities` and `Trusted Publishers` stores using `certlm` Certificate Manager (`All Tasks → Import`).

Microsoft reference: https://learn.microsoft.com/en-us/windows-hardware/drivers/install/introduction-to-test-signing

> [!tip]
> Save the PFX file for build procedures. The same certificate works for signing any number of builds.

## Build Hexagon backend with signed HTP ops libraries

The build procedure is the same as other platforms, but requires additional environment variables:

```
> $env:OPENCL_SDK_ROOT="C:\Qualcomm\OpenCL_SDK\2.3.2"
> $env:HEXAGON_SDK_ROOT="C:\Qualcomm\Hexagon_SDK\6.4.0.2"
> $env:HEXAGON_TOOLS_ROOT="C:\Qualcomm\Hexagon_SDK\6.4.0.2\tools\HEXAGON_Tools\19.0.04"
> $env:HEXAGON_HTP_CERT="c:\Users\MyUsers\Certs\ggml-htp-v1.pfx"
> $env:WINDOWS_SDK_BIN="C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\arm64"

> cmake --preset arm64-windows-snapdragon-release -B build-wos
...
> cmake --install build-wos --prefix pkg-snapdragon
```

Installed HTP ops libraries:

```
> dir pkg-snapdragon/lib
...
-a----         1/22/2026   6:01 PM         187656 libggml-htp-v73.so
-a----         1/22/2026   6:01 PM         191752 libggml-htp-v75.so
-a----         1/22/2026   6:01 PM         187656 libggml-htp-v79.so
-a----         1/22/2026   6:01 PM         187656 libggml-htp-v81.so
-a----         1/22/2026   6:01 PM           4139 libggml-htp.cat
```

Verify the signature:

```
> signtool.exe verify /v /pa .\pkg-snapdragon\lib\libggml-htp.cat
Verifying: .\pkg-snapdragon\lib\libggml-htp.cat

Signature Index: 0 (Primary Signature)
Hash of file (sha256): 9820C664DA59D5EAE31DBB664127FCDAEF59CDC31502496BC567544EC2F401CF

Signing Certificate Chain:
        Issued to: GGML.HTP.v1
...
Successfully verified: .\pkg-snapdragon\lib\libggml-htp.cat
...
```
