package water

import (
	"context"
	"time"

	"aquarecirc/internal/model"
)

func (s *Store) WaitStable(ctx context.Context, pondID string, target float64) error {
	for {
		st, ok := s.State(pondID)
		if !ok {
			return model.ErrNotFound
		}
		if tempStable(st.Temp, target) {
			return nil
		}
		time.Sleep(40 * time.Millisecond)
	}
}

func tempStable(current float64, target float64) bool {
	diff := current - target
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.3
}
