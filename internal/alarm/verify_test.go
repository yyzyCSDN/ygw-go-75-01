package alarm

import (
	"sync"
	"testing"
)

func TestDedupConcurrentSingleAlarm(t *testing.T) {
	bus := NewBus()
	key := NewReportKey("p1", "do", "critical")
	const workers = 24
	results := make([]bool, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = bus.Acquire(key)
		}(i)
	}
	wg.Wait()
	acquired := 0
	for _, ok := range results {
		if ok {
			acquired++
		}
	}
	if acquired != 1 {
		t.Fatalf("expected exactly one acquisition, got %d", acquired)
	}
}
