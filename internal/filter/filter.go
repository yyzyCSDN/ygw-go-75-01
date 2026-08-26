package filter

import (
	"context"

	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

type Filter struct {
	ID      string
	PondID  string
	store   *water.Store
	driver  Driver
	status  *model.DeviceStatus
	enabled bool
}

func New(
	id string,
	pondID string,
	store *water.Store,
	driver Driver,
) *Filter {
	return &Filter{
		ID:      id,
		PondID:  pondID,
		store:   store,
		driver:  driver,
		status:  &model.DeviceStatus{ID: id, Kind: model.DeviceFilter, PondID: pondID},
		enabled: true,
	}
}

func (f *Filter) Status() model.DeviceStatus {
	return *f.status
}

func (f *Filter) AdjustAmmonia(reading float64) error {
	patch := water.PatchAmmonia(reading)
	if err := f.store.ApplyMetrics(f.PondID, patch); err != nil {
		return err
	}
	f.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, f.PondID, "filter adjust", reading))
	return nil
}

func (f *Filter) Start(ctx context.Context) error {
	if !f.enabled {
		return model.ErrUnavailable
	}
	if err := f.driver.Start(ctx); err != nil {
		f.status.Fail(err)
		return err
	}
	f.status.Succeed()
	f.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, f.PondID, "filter start", 1))
	return nil
}

func (f *Filter) Stop(ctx context.Context) error {
	if err := f.driver.Stop(ctx); err != nil {
		f.status.Fail(err)
		return err
	}
	f.status.Running = false
	f.status.Touch()
	f.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, f.PondID, "filter stop", 0))
	return nil
}
