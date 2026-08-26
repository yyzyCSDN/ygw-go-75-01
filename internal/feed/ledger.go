package feed

import (
	"sync"
	"time"
)

type LedgerEntry struct {
	BatchID string
	PondID  string
	Amount  float64
	At      time.Time
}

type Ledger struct {
	mu    sync.Mutex
	items []LedgerEntry
	limit int
}

func NewLedger(limit int) *Ledger {
	return &Ledger{
		items: []LedgerEntry{},
		limit: limit,
	}
}

func (l *Ledger) Append(entry LedgerEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, entry)
	if len(l.items) > l.limit {
		l.items = l.items[len(l.items)-l.limit:]
	}
}

func (l *Ledger) Recent(pondID string, limit int) []LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LedgerEntry, 0, limit)
	for i := len(l.items) - 1; i >= 0 && len(out) < limit; i-- {
		if l.items[i].PondID == pondID {
			out = append(out, l.items[i])
		}
	}
	return out
}
