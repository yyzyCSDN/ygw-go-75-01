package filter

import (
	"context"
	"sync"
	"time"

	"aquarecirc/internal/model"
)

type BackwashStatus struct {
	LastAt     time.Time
	Runtime    time.Duration
	CycleCount int
	Due        bool
}

type backwashTracker struct {
	mu      sync.Mutex
	lastAt  time.Time
	running time.Duration
	cycle   int
}

func (t *backwashTracker) addRuntime(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running += d
}

func (t *backwashTracker) mark(at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastAt = at
	t.cycle++
	t.running = 0
}

func (t *backwashTracker) status() BackwashStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	return BackwashStatus{
		LastAt:     t.lastAt,
		Runtime:    t.running,
		CycleCount: t.cycle,
	}
}

type FilterBackwash struct {
	filter   *Filter
	tracker  *backwashTracker
	interval time.Duration
}

func NewBackwash(filter *Filter, interval time.Duration) *FilterBackwash {
	return &FilterBackwash{
		filter:   filter,
		tracker:  &backwashTracker{},
		interval: interval,
	}
}

func (b *FilterBackwash) Track(d time.Duration) {
	b.tracker.addRuntime(d)
}

func (b *FilterBackwash) Due() bool {
	st := b.tracker.status()
	if st.LastAt.IsZero() {
		return st.Runtime >= b.interval
	}
	return st.Runtime >= b.interval || time.Since(st.LastAt) >= 6*b.interval
}

func (b *FilterBackwash) Run(ctx context.Context) error {
	if err := b.filter.driver.Start(ctx); err != nil {
		return err
	}
	select {
	case <-time.After(8 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := b.filter.driver.Stop(ctx); err != nil {
		return err
	}
	b.tracker.mark(time.Now())
	b.filter.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, b.filter.PondID, "backwash", 1))
	return nil
}

func (b *FilterBackwash) Status() BackwashStatus {
	st := b.tracker.status()
	st.Due = b.Due()
	return st
}
