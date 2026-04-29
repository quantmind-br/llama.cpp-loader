---
title: Debugging tests
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/development/debugging-tests.md
source: git
fetched_at: 2026-04-28T09:49:23.987599099-03:00
rendered_js: false
word_count: 180
summary: This document describes how to efficiently run and debug specific project tests using a dedicated shell script and standard command-line tools like GDB.
tags:
    - debugging
    - testing
    - gdb
    - shell-scripts
    - build-automation
    - development-workflow
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Debugging Tests Tips

Use the `debug-test.sh` script in `scripts/` to run or debug specific tests with a short feedback loop.

## Quick usage

```
debug-test.sh [OPTION]... <test_regex> <test_number>
```

| Command | What it does |
| --- | --- |
| `./scripts/debug-test.sh test-tokenizer` | Execute test, get PASS/FAIL |
| `./scripts/debug-test.sh -g test-tokenizer` | Run in GDB |
| `./scripts/debug-test.sh test 23` | Run specific test number |
| `./scripts/debug-test.sh -h` | Print help |

In GDB, set breakpoints at the prompt:

```bash
>>> b main
```

## How the script works

### Step 1: Reset and setup folder context

```bash
rm -rf build-ci-debug && mkdir build-ci-debug && cd build-ci-debug
```

### Step 2: Build debug test binaries

```bash
cmake -DCMAKE_BUILD_TYPE=Debug -DLLAMA_CUDA=1 -DLLAMA_FATAL_WARNINGS=ON ..
make -j
```

### Step 3: Find tests matching REGEX

| Flag | Purpose |
| --- | --- |
| `-R test-tokenizer` | Match test files named `test-tokenizer*` |
| `-N` | Show-only mode: display test commands without executing |
| `-V` | Verbose output |

```bash
ctest -R "test-tokenizer" -V -N
```

Sample output:

```bash
...
1: Test command: ~/llama.cpp/build-ci-debug/bin/test-tokenizer-0 "~/llama.cpp/tests/../models/ggml-vocab-llama-spm.gguf"
1: Working Directory: .
Labels: main
  Test  #1: test-tokenizer-0-llama-spm
...
4: Test command: ~/llama.cpp/build-ci-debug/bin/test-tokenizer-0 "~/llama.cpp/tests/../models/ggml-vocab-falcon.gguf"
4: Working Directory: .
Labels: main
  Test  #4: test-tokenizer-0-falcon
...
```

### Step 4: Identify the test command

From the output above, for test #1:

- **Test Binary:** `~/llama.cpp/build-ci-debug/bin/test-tokenizer-0`
- **Test GGUF Model:** `~/llama.cpp/tests/../models/ggml-vocab-llama-spm.gguf`

### Step 5: Run GDB on the test command

```bash
gdb --args ~/llama.cpp/build-ci-debug/bin/test-tokenizer-0 "~/llama.cpp/tests/../models/ggml-vocab-llama-spm.gguf"
```

#debugging #testing #gdb #development-workflow