package alarm

import (
	"sync"
	"time"

	"aquarecirc/internal/model"
)

const DedupWindow = 60 * time.Second

type Sink interface {
	Emit(model.Alarm)
}

type Bus struct {
	mu      sync.Mutex
	gate    *ReportGate
	sinks   []Sink
	history []model.Alarm
	counts  map[string]int
}

func NewBus() *Bus {
	return &Bus{
		gate:   NewReportGate(),
		counts: map[string]int{},
	}
}

func (b *Bus) Acquire(key ReportKey) bool {
	return b.gate.Acquire(key, DedupWindow)
}

func (b *Bus) Report(evt model.Alarm) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = append(b.history, evt)
	b.counts[evt.Key()]++
	for _, sink := range b.sinks {
		sink.Emit(evt)
	}
	return nil
}

func (b *Bus) AddSink(sink Sink) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sinks = append(b.sinks, sink)
}

func (b *Bus) Recent(limit int) []model.Alarm {
	b.mu.Lock()
	defer b.mu.Unlock()
	if limit <= 0 || limit > len(b.history) {
		limit = len(b.history)
	}
	out := make([]model.Alarm, limit)
	copy(out, b.history[len(b.history)-limit:])
	return out
}

func (b *Bus) Count(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.counts[key]
}

func (b *Bus) Prune(maxAge time.Duration) int {
	return b.gate.Prune(maxAge)
}

func (b *Bus) GateSize() int {
	return b.gate.SnapshotSize()
}

func (b *Bus) SuppressedTotal() int {
	return b.gate.SuppressedTotal()
}

func (b *Bus) ReportFailure(pondID string, device string, err error) error {
	if err == nil {
		return nil
	}
	return err
}
