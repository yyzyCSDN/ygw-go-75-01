package water

import (
	"sync"
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

func TestWaterQualityConcurrentUpdateNoLoss(t *testing.T) {
	bus := alarm.NewBus()
	store := NewStore(bus)
	if err := store.RegisterPond(model.NewPond("p1", "主池", 100, 1000)); err != nil {
		t.Fatalf("register: %v", err)
	}
	const rounds = 60
	var wg sync.WaitGroup
	for i := 0; i < rounds; i++ {
		do := float64(5 + i%3)
		ammonia := float64(0.2 + float64(i%5)*0.1)
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = store.ApplyMetrics("p1", PatchDO(do))
		}()
		go func() {
			defer wg.Done()
			_ = store.ApplyMetrics("p1", PatchAmmonia(ammonia))
		}()
	}
	wg.Wait()
	st, ok := store.State("p1")
	if !ok {
		t.Fatal("state missing")
	}
	if st.DO <= 0 {
		t.Fatalf("DO lost: %v", st.DO)
	}
	if st.Ammonia <= 0 {
		t.Fatalf("ammonia lost: %v", st.Ammonia)
	}
}
