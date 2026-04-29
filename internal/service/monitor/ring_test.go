package monitor

import (
	"sync"
	"testing"
)

func TestLogRing_PushUnderCapacity(t *testing.T) {
	r := newLogRing(5)
	r.push("a")
	r.push("b")
	r.push("c")
	got := r.snapshot()
	if want := []string{"a", "b", "c"}; !equalSlices(got, want) {
		t.Fatalf("snapshot = %v, want %v", got, want)
	}
}

func TestLogRing_PushOverwritesOldest(t *testing.T) {
	r := newLogRing(3)
	r.push("a")
	r.push("b")
	r.push("c")
	r.push("d") // overwrites "a"
	got := r.snapshot()
	if want := []string{"b", "c", "d"}; !equalSlices(got, want) {
		t.Fatalf("after overflow: snapshot = %v, want %v", got, want)
	}
}

func TestLogRing_ConcurrentPushSafe(t *testing.T) {
	r := newLogRing(1000)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.push("line")
			}
		}()
	}
	wg.Wait()
	if got := len(r.snapshot()); got != 1000 {
		t.Fatalf("snapshot len = %d, want 1000", got)
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
