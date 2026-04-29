package monitor

import "sync"

// logRing is a fixed-capacity, thread-safe ring buffer of strings.
// Once full, push overwrites the oldest entry.
type logRing struct {
	mu   sync.Mutex
	buf  []string
	head int // next write position
	full bool
}

func newLogRing(cap int) *logRing {
	return &logRing{buf: make([]string, cap)}
}

func (r *logRing) push(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = s
	r.head = (r.head + 1) % len(r.buf)
	if r.head == 0 {
		r.full = true
	}
}

// snapshot returns a chronologically-ordered copy.
func (r *logRing) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.full {
		out := make([]string, r.head)
		copy(out, r.buf[:r.head])
		return out
	}
	out := make([]string, len(r.buf))
	copy(out, r.buf[r.head:])
	copy(out[len(r.buf)-r.head:], r.buf[:r.head])
	return out
}
