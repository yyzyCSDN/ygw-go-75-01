package alarm

import (
	"testing"

	"aquarecirc/internal/model"
)

// TestSnapshotCache_PondIsolation reproduces the cross-pond contamination
// bug: writing back pond A and then pond B must not let B overwrite A's
// cached reading. Before the fix, both map slots pointed at the same
// shared backing variable, so A's Ammonia was silently replaced by B's.
func TestSnapshotCache_PondIsolation(t *testing.T) {
	c := NewSnapshotCache()

	pondA := model.WaterSnapshot{PondID: "A", Ammonia: 1.8, DO: 6.0, Level: model.LevelCritical}
	pondB := model.WaterSnapshot{PondID: "B", Ammonia: 0.2, DO: 7.5, Level: model.LevelNormal}

	c.WriteBack("A", &pondA)
	c.WriteBack("B", &pondB)

	gotA, ok := c.Latest("A")
	if !ok {
		t.Fatal("expected pond A snapshot to be cached")
	}
	gotB, ok := c.Latest("B")
	if !ok {
		t.Fatal("expected pond B snapshot to be cached")
	}

	// A must keep its own critical ammonia reading, not be dragged to B's value.
	if gotA.Ammonia != pondA.Ammonia {
		t.Errorf("pond A ammonia contaminated: got %v, want %v", gotA.Ammonia, pondA.Ammonia)
	}
	if gotA.Level != pondA.Level {
		t.Errorf("pond A level contaminated: got %v, want %v", gotA.Level, pondA.Level)
	}
	// B keeps its own reading too.
	if gotB.Ammonia != pondB.Ammonia {
		t.Errorf("pond B ammonia contaminated: got %v, want %v", gotB.Ammonia, pondB.Ammonia)
	}

	// Mutating the caller's source after WriteBack must not bleed into the cache:
	// the cache owns an independent copy, never a pointer into caller memory.
	pondA.Ammonia = 9.9
	gotA2, _ := c.Latest("A")
	if gotA2.Ammonia != 1.8 {
		t.Errorf("cache aliases caller memory: mutating source changed cached value to %v", gotA2.Ammonia)
	}
}

// TestSnapshotCache_OverwritePond verifies re-sampling the same pond
// updates only that pond and leaves others untouched.
func TestSnapshotCache_OverwritePond(t *testing.T) {
	c := NewSnapshotCache()

	c.WriteBack("A", &model.WaterSnapshot{PondID: "A", Ammonia: 0.3})
	c.WriteBack("B", &model.WaterSnapshot{PondID: "B", Ammonia: 0.4})
	// Re-sample A with a new value.
	c.WriteBack("A", &model.WaterSnapshot{PondID: "A", Ammonia: 2.1})

	gotA, _ := c.Latest("A")
	gotB, _ := c.Latest("B")

	if gotA.Ammonia != 2.1 {
		t.Errorf("pond A not updated: got %v, want 2.1", gotA.Ammonia)
	}
	if gotB.Ammonia != 0.4 {
		t.Errorf("pond B contaminated by A re-sample: got %v, want 0.4", gotB.Ammonia)
	}
}

func TestSnapshotCache_NilAndMissing(t *testing.T) {
	c := NewSnapshotCache()

	c.WriteBack("A", nil) // must be a no-op, not a panic
	if _, ok := c.Latest("A"); ok {
		t.Error("nil write should not populate the cache")
	}
	if _, ok := c.Latest("missing"); ok {
		t.Error("missing pond should report not-ok")
	}
	if size := c.Size(); size != 0 {
		t.Errorf("expected empty cache, got size %d", size)
	}
}
