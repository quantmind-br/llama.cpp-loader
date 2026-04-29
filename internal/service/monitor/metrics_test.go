package monitor

import (
	"testing"
	"time"
)

func TestMetricsAgg_TokensPerSecFromLog(t *testing.T) {
	a := newMetricsAgg(60 * time.Second)
	now := time.Unix(1_700_000_000, 0)

	// llama-server typically logs:
	//   prompt eval time =     14.36 ms /     5 tokens (    2.87 ms per token,   348.40 tokens per second)
	a.observeLog(now, "prompt eval time = 14.36 ms / 5 tokens ( 2.87 ms per token,   348.40 tokens per second)")
	a.observeLog(now.Add(time.Second), "eval time = 100.00 ms / 50 tokens ( 2.00 ms per token,   500.00 tokens per second)")

	m := a.snapshot(now.Add(time.Second))
	if got := m.TokensPerSec[len(m.TokensPerSec)-1]; got != 500 {
		t.Fatalf("latest tokens/s = %v, want 500", got)
	}
}

func TestMetricsAgg_RequestsPerSecFromSlotsDiff(t *testing.T) {
	a := newMetricsAgg(60 * time.Second)
	now := time.Unix(1_700_000_000, 0)

	a.observeSlots(now, SlotSnapshot{Slots: []Slot{{ID: 0, NDecoded: 100}}})
	a.observeSlots(now.Add(time.Second), SlotSnapshot{Slots: []Slot{{ID: 0, NDecoded: 110}}})
	a.observeSlots(now.Add(2*time.Second), SlotSnapshot{Slots: []Slot{{ID: 0, NDecoded: 130}}})

	m := a.snapshot(now.Add(2 * time.Second))
	// Two diffs: 10/s then 20/s; we look at the most-recent bucket.
	if got := m.RequestsPerSec[len(m.RequestsPerSec)-1]; got != 20 {
		t.Fatalf("latest req/s = %v, want 20", got)
	}
}

func TestMetricsAgg_DropsOldSamples(t *testing.T) {
	a := newMetricsAgg(2 * time.Second)
	t0 := time.Unix(1_700_000_000, 0)
	a.observeLog(t0, "tokens per second 100.0")
	a.observeLog(t0.Add(3*time.Second), "tokens per second 200.0")

	m := a.snapshot(t0.Add(3 * time.Second))
	// 100.0 sample is older than window — must be excluded.
	for _, v := range m.TokensPerSec {
		if v == 100 {
			t.Fatalf("stale 100.0 sample retained: %v", m.TokensPerSec)
		}
	}
}
