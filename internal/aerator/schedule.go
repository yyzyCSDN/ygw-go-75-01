package aerator

import (
	"context"
	"time"

	"aquarecirc/internal/model"
)

func (a *Aerator) RunAuto(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !a.retryAt.IsZero() && a.retryAt.After(time.Now()) {
				continue
			}
			a.tick()
		}
	}
}

func (a *Aerator) tick() {
	st, ok := a.store.State(a.PondID)
	if !ok {
		return
	}
	if st.Level == model.LevelCritical && !a.status.Running {
		_ = a.Start(context.Background())
		return
	}
	if st.Level != model.LevelCritical && a.status.Running && st.DO >= a.pondTarget() {
		_ = a.Stop(context.Background())
	}
}

func (a *Aerator) pondTarget() float64 {
	pond, ok := a.store.Pond(a.PondID)
	if !ok {
		return 6
	}
	return pond.TargetDO
}
