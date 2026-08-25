package water

import (
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

func TestPondCacheIsolatedPerPond(t *testing.T) {
	bus := alarm.NewBus()
	store := NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	_ = store.RegisterPond(model.NewPond("p2", "B", 80, 600))
	snapA, ok := store.TakeSnapshot("p1")
	if !ok {
		t.Fatal("snapshot p1 missing")
	}
	snapA.DO = 7.5
	store.WriteBackSnapshot("p1", snapA)
	snapB, ok := store.TakeSnapshot("p2")
	if !ok {
		t.Fatal("snapshot p2 missing")
	}
	snapB.DO = 5.1
	store.WriteBackSnapshot("p2", snapB)
	latest, ok := store.LatestSnapshot("p1")
	if !ok {
		t.Fatal("p1 latest missing")
	}
	if latest.DO != 7.5 {
		t.Fatalf("p1 snapshot polluted by p2: got %v want 7.5", latest.DO)
	}
}
