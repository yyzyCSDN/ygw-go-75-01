package water

import (
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

// TestStore_SampleAll_PondIsolation reproduces the reported production bug:
// after sampling every pond, each pond's cached snapshot must reflect only
// its own readings. Before the fix, A and B shared a single backing slot,
// so the last pond sampled dragged every earlier pond's ammonia to its value.
func TestStore_SampleAll_PondIsolation(t *testing.T) {
	store := NewStore(alarm.NewBus())

	pondA := model.NewPond("A", "Pond A", 100, 1000)
	pondB := model.NewPond("B", "Pond B", 100, 1000)
	if err := store.RegisterPond(pondA); err != nil {
		t.Fatalf("register A: %v", err)
	}
	if err := store.RegisterPond(pondB); err != nil {
		t.Fatalf("register B: %v", err)
	}

	// Push pond A into a critical ammonia state, leave B normal.
	if err := store.ApplyMetrics("A", PatchAmmonia(1.8)); err != nil {
		t.Fatalf("apply A: %v", err)
	}
	if err := store.ApplyMetrics("B", PatchAmmonia(0.2)); err != nil {
		t.Fatalf("apply B: %v", err)
	}

	store.SampleAll()

	gotA, ok := store.LatestSnapshot("A")
	if !ok {
		t.Fatal("expected pond A snapshot")
	}
	gotB, ok := store.LatestSnapshot("B")
	if !ok {
		t.Fatal("expected pond B snapshot")
	}

	// A must retain its own critical reading and not be dragged to B's value.
	if gotA.Ammonia != 1.8 {
		t.Errorf("pond A ammonia contaminated by B: got %v, want 1.8", gotA.Ammonia)
	}
	if gotA.Level != model.LevelCritical {
		t.Errorf("pond A level lost: got %v, want critical", gotA.Level)
	}
	if gotB.Ammonia != 0.2 {
		t.Errorf("pond B ammonia contaminated by A: got %v, want 0.2", gotB.Ammonia)
	}
	if gotB.PondID != "B" {
		t.Errorf("pond B PondID contaminated: got %q, want \"B\"", gotB.PondID)
	}
}

// TestStore_WriteBackSnapshot_NoCallerAlias ensures the cache does not alias
// the caller's snapshot value. Mutating the source after WriteBack must not
// bleed into the cached reading.
func TestStore_WriteBackSnapshot_NoCallerAlias(t *testing.T) {
	store := NewStore(alarm.NewBus())
	if err := store.RegisterPond(model.NewPond("A", "Pond A", 100, 1000)); err != nil {
		t.Fatal(err)
	}

	snap := model.WaterSnapshot{PondID: "A", Ammonia: 0.5, DO: 6.0}
	store.WriteBackSnapshot("A", snap)

	snap.Ammonia = 9.9 // mutate caller's local copy

	got, _ := store.LatestSnapshot("A")
	if got.Ammonia != 0.5 {
		t.Errorf("cache aliases caller memory: got %v, want 0.5", got.Ammonia)
	}
}
