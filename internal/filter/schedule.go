package filter

import (
	"context"
	"time"
)

func (f *Filter) RunAuto(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.tick()
		}
	}
}

func (f *Filter) tick() {
	st, ok := f.store.State(f.PondID)
	if !ok {
		return
	}
	if st.Ammonia > 0.7 && !f.status.Running {
		_ = f.Start(context.Background())
		return
	}
	if st.Ammonia <= 0.4 && f.status.Running {
		_ = f.Stop(context.Background())
	}
}
