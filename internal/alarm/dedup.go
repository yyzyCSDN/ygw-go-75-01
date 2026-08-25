package alarm

import (
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
)

type ReportKey struct {
	PondID string
	Metric string
	Level  string
}

func NewReportKey(pondID string, metric string, level string) ReportKey {
	return ReportKey{
		PondID: pondID,
		Metric: metric,
		Level:  level,
	}
}

func KeyHash(key ReportKey) uint64 {
	hasher := xxhash.New()
	hasher.WriteString(key.PondID)
	hasher.WriteString("|")
	hasher.WriteString(key.Metric)
	hasher.WriteString("|")
	hasher.WriteString(key.Level)
	return hasher.Sum64()
}

type seenEntry struct {
	acquiredAt time.Time
	count      int
	hash       uint64
}

type ReportGate struct {
	mu    sync.Mutex
	seen  map[ReportKey]seenEntry
	order []ReportKey
	now   func() time.Time
}

func NewReportGate() *ReportGate {
	return &ReportGate{
		seen: map[ReportKey]seenEntry{},
		now:  time.Now,
	}
}

func (g *ReportGate) Acquire(key ReportKey, window time.Duration) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	entry, ok := g.seen[key]
	if ok && now.Sub(entry.acquiredAt) < window {
		entry.count++
		g.seen[key] = entry
		return false
	}
	if !ok {
		g.order = append(g.order, key)
	}
	g.seen[key] = seenEntry{acquiredAt: now, count: 1, hash: KeyHash(key)}
	if len(g.order) > 256 {
		oldest := g.order[0]
		g.order = g.order[1:]
		delete(g.seen, oldest)
	}
	return true
}

func (g *ReportGate) Prune(maxAge time.Duration) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	cutoff := g.now().Add(-maxAge)
	removed := 0
	for key, entry := range g.seen {
		if entry.acquiredAt.Before(cutoff) {
			delete(g.seen, key)
			removed++
		}
	}
	return removed
}

func (g *ReportGate) SnapshotSize() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.seen)
}

func (g *ReportGate) SuppressedTotal() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	total := 0
	for key, entry := range g.seen {
		if entry.hash != KeyHash(key) {
			continue
		}
		total += entry.count - 1
	}
	return total
}
