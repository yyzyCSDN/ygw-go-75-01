package temp

import (
	"context"

	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

type Controller struct {
	store *water.Store
	state map[string]model.TempRegulation
}

func NewController(store *water.Store) *Controller {
	return &Controller{
		store: store,
		state: map[string]model.TempRegulation{},
	}
}

func (c *Controller) Regulate(ctx context.Context, pondID string, target float64) error {
	reg := model.NewTempRegulation(pondID, target)
	reg.Current = c.currentTemp(pondID)
	c.state[pondID] = reg
	if err := c.store.WaitStable(ctx, pondID, target); err != nil {
		reg.Abort()
		c.state[pondID] = reg
		c.store.AddEvent(model.NewTelemetry(model.TelemetryTemp, pondID, "temp aborted", target))
		return err
	}
	st, ok := c.store.State(pondID)
	if !ok {
		reg.Abort()
		c.state[pondID] = reg
		return model.ErrNotFound
	}
	reg.Stabilize(st.Temp)
	c.state[pondID] = reg
	c.store.AddEvent(model.NewTelemetry(model.TelemetryTemp, pondID, "temp stable", target))
	return nil
}

func (c *Controller) currentTemp(pondID string) float64 {
	st, ok := c.store.State(pondID)
	if !ok {
		return 0
	}
	return st.Temp
}

func (c *Controller) Status(pondID string) (model.TempRegulation, bool) {
	reg, ok := c.state[pondID]
	return reg, ok
}
