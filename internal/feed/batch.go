package feed

import (
	"context"
	"sync"

	"aquarecirc/internal/model"
)

type Scheduler struct {
	feeder  *Feeder
	mu      sync.Mutex
	running map[string]bool
}

func NewScheduler(feeder *Feeder) *Scheduler {
	return &Scheduler{
		feeder:  feeder,
		running: map[string]bool{},
	}
}

func (s *Scheduler) Start(ctx context.Context, pondID string) (*model.FeedBatch, error) {
	s.mu.Lock()
	if s.running[pondID] {
		s.mu.Unlock()
		return nil, model.ErrBusy
	}
	s.running[pondID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running[pondID] = false
		s.mu.Unlock()
	}()
	batch, err := s.feeder.Plan(pondID)
	if err != nil {
		return nil, err
	}
	err = s.feeder.Run(ctx, batch)
	return batch, err
}
