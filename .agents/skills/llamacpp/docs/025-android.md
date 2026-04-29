---
title: Android
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/android.md
source: git
fetched_at: 2026-04-28T09:49:08.915697723-03:00
rendered_js: false
word_count: 404
summary: 'This document provides instructions for deploying and building llama.cpp on Android devices using three different methods: Android Studio integration, Termux terminal emulation, and Android NDK cross-compilation.'
tags:
    - android-development
    - llama-cpp
    - cross-compilation
    - mobile-inference
    - termux
    - ndk
    - gguf
category: guide
optimized: true
optimized_at: '2026-04-28T00:00:00Z'
---
# Android

## Build GUI binding using Android Studio

Import the `examples/llama.android` directory into Android Studio, then perform a Gradle sync and build the project.
![Project imported into Android Studio](./android/imported-into-android-studio.jpg)

This Android binding supports hardware acceleration up to `SME2` for **Arm** and `AMX` for **x86-64** CPUs on Android and ChromeOS devices. It auto-detects host hardware to load compatible kernels, running seamlessly on both latest premium devices and older devices without manual configuration.

A minimal Android app frontend is included to showcase core functionalities:
1. **Parse GGUF metadata** via `GgufMetadataReader` from a `ContentResolver` `Uri` (shared storage) or a local `File` (app private storage).
2. **Obtain an `InferenceEngine`** instance through the `AiChat` facade and load a model via its app-private file path.
3. **Send a raw user prompt** for automatic template formatting, prefill, and batch decoding. Collect generated tokens in a Kotlin `Flow`.

For a production-ready experience with system prompts, benchmarks, model management, and Arm feature visualizer, see [Arm AI Chat](https://play.google.com/store/apps/details?id=com.arm.aichat) on Google Play (by Arm's **CT-ML**, **CE-ML** and **STE** groups):

| ![Home screen](https://naco-siren.github.io/ai-chat/policy/index/1-llm-starter-pack.png)  | ![System prompt](https://naco-siren.github.io/ai-chat/policy/index/5-system-prompt.png)  | !["Haiku"](https://naco-siren.github.io/ai-chat/policy/index/4-metrics.png)  |
|:------------------------------------------------------:|:----------------------------------------------------:|:--------------------------------------------------------:|
|                      Home screen                       |                    System prompt                     |                         "Haiku"                          |

## Build CLI on Android using Termux

[Termux](https://termux.dev/en/) is an Android terminal emulator and Linux environment app (no root required). Available experimentally in the Google Play Store, or directly from the project repo or F-Droid.

With Termux, install and run `llama.cpp` as on Linux:

```
$ apt update && apt upgrade -y
$ apt install git cmake
```

Then follow the [build instructions](027-build.md) for CMake.

Once built, download a model (e.g., from Hugging Face) to `~/` for best performance:

```
$ curl -L {model-url} -o ~/{model}.gguf
```

Run inference:

```
$ ./build/bin/llama-cli -m ~/{model}.gguf -c {context-size} -p "{your-prompt}"
```

Set `context-size` to a reasonable number (e.g., 4096) to avoid memory spikes.

Demo running on a Pixel 5:

https://user-images.githubusercontent.com/271616/225014776-1d567049-ad71-4ef2-b050-55b0b3b9274c.mp4

## Cross-compile CLI using Android NDK

Build `llama.cpp` for Android on your host system via CMake and the Android NDK. Ensure the Android SDK is installed. Note: Android ships with a limited set of native libraries (see: https://developer.android.com/ndk/guides/stable_apis).

After cloning `llama.cpp`:

```
$ cmake \
  -DCMAKE_TOOLCHAIN_FILE=$ANDROID_NDK/build/cmake/android.toolchain.cmake \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-28 \
  -DCMAKE_C_FLAGS="-march=armv8.7a" \
  -DCMAKE_CXX_FLAGS="-march=armv8.7a" \
  -DGGML_OPENMP=OFF \
  -DGGML_LLAMAFILE=OFF \
  -B build-android
```

- OpenMP must be installed by CMake as a dependency (not supported on Android at this time)
- `llamafile` does not support Android devices (see: https://github.com/Mozilla-Ocho/llamafile/issues/325)

The above configures with the most performant options for modern devices. Runtime CPU feature checks handle devices not running `armv8.7a`. Adjust `ANDROID_ABI` as needed.

Build and install:

```
$ cmake --build build-android --config Release -j{n}
$ cmake --install build-android --prefix {install-dir} --config Release
```

Push to device and run:

```
$ adb shell "mkdir /data/local/tmp/llama.cpp"
$ adb push {install-dir} /data/local/tmp/llama.cpp/
$ adb push {model}.gguf /data/local/tmp/llama.cpp/
$ adb shell
```

```
$ cd /data/local/tmp/llama.cpp
$ LD_LIBRARY_PATH=lib ./bin/llama-simple -m {model}.gguf -c {context-size} -p "{your-prompt}"
```

> [!warning]
> Android won't find the library path `lib` on its own — `LD_LIBRARY_PATH` is required. Android supports `RPATH` in later API levels, so this may change in the future.
