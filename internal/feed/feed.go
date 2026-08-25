package feed

import (
	"context"
	"fmt"
	"time"

	"aquarecirc/internal/model"
	"aquarecirc/internal/pump"
	"aquarecirc/internal/water"
)

type Recorder struct {
	store *water.Store
}

func NewRecorder(store *water.Store) *Recorder {
	return &Recorder{store: store}
}

func (r *Recorder) Remaining(pondID string) float64 {
	return r.store.FeedBalance(pondID)
}

func (r *Recorder) Commit(batch *model.FeedBatch) error {
	return r.store.DeductBalance(batch.PondID, batch.UsedAmount)
}

type Feeder struct {
	store    *water.Store
	pump     *pump.Pump
	recorder *Recorder
	ledger   *Ledger
	seq      int
}

func NewFeeder(store *water.Store, pump *pump.Pump) *Feeder {
	return &Feeder{
		store:    store,
		pump:     pump,
		recorder: NewRecorder(store),
		ledger:   NewLedger(200),
	}
}

func (f *Feeder) Ledger() *Ledger {
	return f.ledger
}

func (f *Feeder) Plan(pondID string) (*model.FeedBatch, error) {
	pond, ok := f.store.Pond(pondID)
	if !ok {
		return nil, model.ErrNotFound
	}
	remaining := f.recorder.Remaining(pondID)
	planned := pond.FeedAmount()
	if planned > remaining {
		planned = remaining
	}
	f.seq++
	id := fmt.Sprintf("F-%s-%d", pondID, f.seq)
	return model.NewFeedBatch(id, pondID, planned), nil
}

func (f *Feeder) Run(ctx context.Context, batch *model.FeedBatch) error {
	if batch.Status != model.FeedIdle {
		return model.ErrBusy
	}
	batch.Status = model.FeedFeeding
	if err := f.pump.BeginFeeding(); err != nil {
		batch.Finish(0, err.Error())
		return err
	}
	batch.Status = model.FeedPausing
	defer func() {
		_ = f.pump.EndFeeding()
	}()
	used := batch.PlannedAmount
	select {
	case <-time.After(15 * time.Millisecond):
	case <-ctx.Done():
		batch.Finish(0, ctx.Err().Error())
		return ctx.Err()
	}
	batch.Finish(used, "completed")
	if err := f.recorder.Commit(batch); err != nil {
		return err
	}
	f.ledger.Append(LedgerEntry{
		BatchID: batch.ID,
		PondID:  batch.PondID,
		Amount:  used,
		At:      batch.FinishedAt,
	})
	return nil
}
