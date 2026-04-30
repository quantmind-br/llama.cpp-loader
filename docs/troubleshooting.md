# Troubleshooting

Common issues and their solutions.

## llama-server not found

**Symptom:** Status bar shows "llama-server binary not found in PATH" at startup.

**Solution:**
- Ensure you have compiled [llama.cpp](https://github.com/ggerganov/llama.cpp) with server support
- Verify `llama-server` is in your `PATH`:
  ```bash
  which llama-server
  ```
- If installed in a custom location, add it to your shell profile:
  ```bash
  export PATH="/path/to/llama.cpp/build/bin:$PATH"
  ```

## Port already in use

**Symptom:** Launch fails with "port already in use".

**Solution:**
- Edit the profile and change the `port` argument
- Or kill the existing instance from the Monitor tab (`k`)
- Use the Launcher tab to check if an instance is already running on that port

## Model file not found

**Symptom:** Launch fails with "model file not found".

**Solution:**
- Verify the model path in the profile editor
- Ensure the file exists and is readable
- Update `search_paths` in `config.toml` to include the directory containing your models:
  ```toml
  [models]
  search_paths = ["~/models", "/path/to/your/models"]
  ```

## Background instance not recovered

**Symptom:** A background instance launched before closing the TUI does not appear in the Monitor tab on restart.

**Solution:**
- Check that `instances.json` exists in your state directory (`~/.local/state/llama-cpp-loader/`)
- Verify the process is still running: `ps aux | grep llama-server`
- If the process crashed, it will be marked with a crash indicator; clear it from the Monitor tab

## Foreground instance already running

**Symptom:** Launch fails with "a foreground instance is already running".

**Solution:**
- Only one foreground instance is allowed at a time
- Switch to background mode in the Launcher tab (`b`) before launching
- Or kill the existing foreground instance from the Monitor tab

## Health check timeout

**Symptom:** Launch succeeds but status shows "did not become healthy within timeout".

**Solution:**
- The instance may be taking longer than expected to load the model
- Check the logs in the Monitor tab for errors
- Verify the model file is valid and compatible with your llama-server version
- Large models may require a longer timeout; this is not currently configurable

## Model browser shows no files

**Symptom:** Models tab is empty.

**Solution:**
- Check that `search_paths` in `config.toml` points to directories containing `.gguf` files
- Ensure the directories are readable
- Wait for the scan to complete; large directories may take a few seconds

## Validation errors when saving a profile

**Symptom:** Profile editor shows red validation errors.

**Solution:**
- Validation is version-aware: it checks flags against the `llama-server --help` output of your installed binary
- If `llama-server` is not in PATH, an embedded fallback schema (v7376) is used
- Update llama-server to the latest version if you need newer flags
- Use the Advanced tab to enter flags not available in the Essentials tab

## Clipboard not working

**Symptom:** Copying model path (`c` in Models tab) does nothing.

**Solution:**
- The clipboard integration requires a display server (X11 or Wayland)
- On headless systems, clipboard operations are not available
- The model path is also shown in the UI detail panel for manual copying
