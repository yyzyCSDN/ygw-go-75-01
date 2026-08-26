package feed

import (
	"context"
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
	"aquarecirc/internal/pump"
	"aquarecirc/internal/water"
)

func TestFeedRecordUpdatedPerBatch(t *testing.T) {
	bus := alarm.NewBus()
	store := water.NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	unit := pump.New("pmp-1", "p1", store, pump.NewSimDriver(), bus)
	feeder := NewFeeder(store, unit)
	batch, err := feeder.Plan("p1")
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := feeder.Run(context.Background(), batch); err != nil {
		t.Fatalf("run: %v", err)
	}
	if balance := store.FeedBalance("p1"); balance >= 100 {
		t.Fatalf("feed record not synced: balance %v", balance)
	}
}
