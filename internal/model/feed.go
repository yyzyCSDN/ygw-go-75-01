package model

import "time"

type FeedStatus int

const (
	FeedIdle FeedStatus = iota
	FeedFeeding
	FeedPausing
	FeedDone
)

type FeedBatch struct {
	ID             string
	PondID         string
	PlannedAmount  float64
	UsedAmount     float64
	RemainingAfter float64
	Status         FeedStatus
	StartedAt      time.Time
	FinishedAt     time.Time
	Reason         string
}

func NewFeedBatch(id string, pondID string, planned float64) *FeedBatch {
	return &FeedBatch{
		ID:            id,
		PondID:        pondID,
		PlannedAmount: planned,
		Status:        FeedIdle,
		StartedAt:     time.Now(),
	}
}

func (b *FeedBatch) Finish(used float64, reason string) {
	b.UsedAmount = used
	b.RemainingAfter = b.PlannedAmount - used
	b.Reason = reason
	b.Status = FeedDone
	b.FinishedAt = time.Now()
}
