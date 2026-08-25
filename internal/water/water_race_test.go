package water

import (
	"sync"
	"testing"

	"aquarecirc/internal/model"
)

// TestApplyMetrics_ConcurrentPatchesPreserveAllMetrics guards the data-loss race
// where two goroutines (the aerator writing DO, the filter writing Ammonia on the
// same pond) each dropped the other's metric because ApplyMetrics was not
// serialized and the patch was applied as a full struct replace rather than a merge.
//
// Before the fix, the slower writer's `*st = next` zeroed every field it did not
// set, so the monitor panel flipped between showing only DO and only Ammonia.
func TestApplyMetrics_ConcurrentPatchesPreserveAllMetrics(t *testing.T) {
	store := NewStore(nil)
	pond := &model.Pond{ID: "pond-1", Name: "koi", Volume: 1000, Stock: 10, TargetDO: 6, TargetTemp: 26}
	if err := store.RegisterPond(pond); err != nil {
		t.Fatalf("register: %v", err)
	}

	const rounds = 400
	var wg sync.WaitGroup
	wg.Add(2)

	// Aerator goroutine: writes DO only.
	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			if err := store.ApplyMetrics("pond-1", PatchDO(7.5)); err != nil {
				t.Errorf("aerator apply: %v", err)
				return
			}
		}
	}()

	// Filter goroutine: writes Ammonia only.
	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			if err := store.ApplyMetrics("pond-1", PatchAmmonia(0.3)); err != nil {
				t.Errorf("filter apply: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	st, ok := store.State("pond-1")
	if !ok {
		t.Fatalf("state missing")
	}
	// Each device only reports its own metric, so both must survive concurrent writes:
	// DO set by the aerator, Ammonia set by the filter, and the untouched defaults
	// (Nitrite/PH/Temp) must not be zeroed out by a full-replace patch.
	if st.DO != 7.5 {
		t.Fatalf("DO not preserved after concurrent writes: got %v, want 7.5", st.DO)
	}
	if st.Ammonia != 0.3 {
		t.Fatalf("Ammonia not preserved after concurrent writes: got %v, want 0.3", st.Ammonia)
	}
	if st.PH == 0 || st.Temp == 0 || st.Nitrite == 0 {
		t.Fatalf("untouched metrics zeroed by patch merge: %+v", st)
	}
}

// TestApplyMetrics_PatchOnlyTouchesOwnedFields verifies the merge semantics in
// isolation: applying a DO-only patch must leave every other metric at its
// existing value, not reset the struct to defaults.
func TestApplyMetrics_PatchOnlyTouchesOwnedFields(t *testing.T) {
	store := NewStore(nil)
	pond := &model.Pond{ID: "pond-2", Name: "koi", Volume: 1000, Stock: 10, TargetDO: 6, TargetTemp: 26}
	if err := store.RegisterPond(pond); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := store.ApplyMetrics("pond-2", PatchAmmonia(0.9)); err != nil {
		t.Fatalf("apply ammonia: %v", err)
	}

	st, ok := store.State("pond-2")
	if !ok {
		t.Fatalf("state missing")
	}
	if st.Ammonia != 0.9 {
		t.Fatalf("Ammonia not applied: got %v want 0.9", st.Ammonia)
	}
	// Defaults from NewWaterState must survive an Ammonia-only patch.
	if st.DO != 6.2 || st.Nitrite != 0.05 || st.PH != 7.0 || st.Temp != 26 {
		t.Fatalf("unrelated metrics clobbered by Ammonia-only patch: %+v", st)
	}
}
