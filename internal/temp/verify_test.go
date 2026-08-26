package temp

import (
	"context"
	"errors"
	"testing"
	"time"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

func TestTempWaitTimeoutCancelled(t *testing.T) {
	bus := alarm.NewBus()
	store := water.NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	controller := NewController(store)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- controller.Regulate(ctx, "p1", 40)
	}()
	select {
	case err := <-done:
		if !errors.Is(err, model.ErrTimeout) {
			t.Fatalf("want timeout error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("regulate did not return after context timeout")
	}
	reg, ok := controller.Status("p1")
	if !ok {
		t.Fatal("regulation status missing")
	}
	if reg.Status != model.TempAborted {
		t.Fatalf("regulation status %v, want aborted", reg.Status)
	}
}
