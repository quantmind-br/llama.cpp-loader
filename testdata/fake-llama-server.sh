#!/usr/bin/env bash
# Fake llama-server for processmgr tests.
#
# Behavior:
#   - Parses --port <N> from args (default 0 = pick free port).
#   - Starts a minimal HTTP server on 127.0.0.1:<N> that responds 200
#     on GET /health with body '{"status":"ok"}'.
#   - Echoes argv (one arg per line, prefixed "arg: ") to stderr.
#   - Stays in foreground until SIGTERM/SIGINT. On signal, exits 0.
#
# Used to exercise Launch/WaitHealthy/Kill/List/Recover without a real
# llama.cpp build. Linux-only; uses /dev/tcp + nc -l for the listener.

set -euo pipefail

PORT=0
ARGS=("$@")
for ((i=0; i<${#ARGS[@]}; i++)); do
  if [[ "${ARGS[i]}" == "--port" ]]; then
    PORT="${ARGS[i+1]:-0}"
  fi
  echo "arg: ${ARGS[i]}" 1>&2
done

if [[ "$PORT" == "0" ]]; then
  echo "fake-llama-server: --port required" 1>&2
  exit 2
fi

# Trap so 'kill' from the test exits cleanly.
cleanup() { exit 0; }
trap cleanup TERM INT

# Tiny HTTP loop: read one request line, reply 200 with JSON, loop forever.
RESPONSE=$'HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 15\r\nConnection: close\r\n\r\n{"status":"ok"}'

# Use python3 if available (most reliable). Falls back to a no-op sleep that
# never serves health (test will time out and surface that).
if command -v python3 >/dev/null 2>&1; then
  exec python3 - "$PORT" <<'PY'
import sys, http.server, socketserver, threading, signal, os
port = int(sys.argv[1])
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            body = b'{"status":"ok"}'
            self.send_response(200); self.send_header("Content-Type","application/json")
            self.send_header("Content-Length", str(len(body))); self.end_headers()
            self.wfile.write(body)
        elif self.path == "/slots":
            body = b'[{"id":0,"state":"idle","n_ctx":0,"n_ctx_total":4096,"n_decoded":0,"n_prompt":0,"id_task":""}]'
            self.send_response(200); self.send_header("Content-Type","application/json")
            self.send_header("Content-Length", str(len(body))); self.end_headers()
            self.wfile.write(body)
        else:
            self.send_response(404); self.end_headers()
    def log_message(self, *a, **kw): pass
srv = socketserver.TCPServer(("127.0.0.1", port), H)
def stop(*_): os._exit(0)
signal.signal(signal.SIGTERM, stop); signal.signal(signal.SIGINT, stop)
t = threading.Thread(target=srv.serve_forever, daemon=True)
t.start()
signal.pause()
PY
fi

# Fallback: just sleep so process stays alive but health never comes up.
while true; do sleep 1; done
