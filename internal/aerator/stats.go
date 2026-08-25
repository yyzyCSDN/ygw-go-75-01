package aerator

import (
	"sync"
	"time"
)

type Stats struct {
	StartCount int
	StopCount  int
	Runtime    time.Duration
	LastStart  time.Time
	LastStop   time.Time
}

type statTracker struct {
	mu        sync.Mutex
	startedAt time.Time
	running   bool
	stats     Stats
}

func (t *statTracker) onStart(at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stats.StartCount++
	t.stats.LastStart = at
	if !t.running {
		t.startedAt = at
		t.running = true
	}
}

func (t *statTracker) onStop(at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stats.StopCount++
	t.stats.LastStop = at
	if t.running {
		t.stats.Runtime += at.Sub(t.startedAt)
		t.running = false
	}
}

func (t *statTracker) snapshot() Stats {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := t.stats
	if t.running {
		out.Runtime += time.Since(t.startedAt)
	}
	return out
}
