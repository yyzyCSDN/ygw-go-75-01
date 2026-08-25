package alarm

import (
	"sync"
	"testing"
	"time"
)

// 并发为同一池同一指标上报,去重窗口内只能放行一次。
// 修复前 Acquire 无锁,check-then-act 会被两个 goroutine 同时通过,
// 导致放行数 > 1 且对 map 并发写。此测试在 -race 下也会跑过。
func TestReportGate_AcquireConcurrentSinglePass(t *testing.T) {
	const workers = 64
	gate := NewReportGate()
	key := NewReportKey("pond-7", "do", "critical")
	window := 60 * time.Second

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		passed   int
		failedAt time.Time
	)
	wg.Add(workers)
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			ok := gate.Acquire(key, window)
			mu.Lock()
			if ok {
				passed++
				if passed > 1 {
					failedAt = time.Now() // 仅记录,race/断言统一在下面
				}
			}
			mu.Unlock()
		}()
	}
	close(start) // 尽量让所有 worker 同时进 Acquire
	wg.Wait()

	if passed != 1 {
		t.Fatalf("expected exactly 1 pass in dedup window, got %d", passed)
	}
	if gate.SuppressedTotal() != workers-1 {
		t.Fatalf("expected %d suppressed, got %d", workers-1, gate.SuppressedTotal())
	}
	_ = failedAt
}

// 窗口外可以再次放行。
func TestReportGate_AcquireReleasesAfterWindow(t *testing.T) {
	gate := NewReportGate()
	key := NewReportKey("pond-7", "do", "critical")

	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	now := t0
	gate.now = func() time.Time { return now }

	if !gate.Acquire(key, 60*time.Second) {
		t.Fatal("first acquire should pass")
	}
	if gate.Acquire(key, 60*time.Second) {
		t.Fatal("second acquire within window should be suppressed")
	}
	now = now.Add(61 * time.Second)
	if !gate.Acquire(key, 60*time.Second) {
		t.Fatal("acquire after window should pass again")
	}
}

// Seen 与 Acquire 视图一致,且自身并发安全。
func TestReportGate_SeenConcurrent(t *testing.T) {
	gate := NewReportGate()
	key := NewReportKey("pond-9", "ammonia", "warning")
	window := 60 * time.Second

	if !gate.Acquire(key, window) {
		t.Fatal("first acquire should pass")
	}

	var wg sync.WaitGroup
	wg.Add(32)
	for i := 0; i < 32; i++ {
		go func() {
			defer wg.Done()
			if !gate.Seen(key) {
				t.Errorf("Seen should be true after acquire")
			}
		}()
	}
	wg.Wait()
}
