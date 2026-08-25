package water

import (
	"context"
	"time"

	"aquarecirc/internal/model"
)

func (s *Store) WaitStable(ctx context.Context, pondID string, target float64) error {
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.record(model.NewTelemetry(model.TelemetryTemp, pondID, "wait cancelled", 0))
			return model.ErrTimeout
		case <-ticker.C:
			st, ok := s.State(pondID)
			if !ok {
				return model.ErrNotFound
			}
			if tempStable(st.Temp, target) {
				return nil
			}
		}
	}
}

func tempStable(current float64, target float64) bool {
	diff := current - target
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.3
}
