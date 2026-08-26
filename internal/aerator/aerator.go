package aerator

import (
	"context"
	"time"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

type Aerator struct {
	ID      string
	PondID  string
	store   *water.Store
	bus     *alarm.Bus
	driver  Driver
	status  *model.DeviceStatus
	enabled bool
	tracker *statTracker
	retryAt time.Time
}

func New(
	id string,
	pondID string,
	store *water.Store,
	bus *alarm.Bus,
	driver Driver,
) *Aerator {
	return &Aerator{
		ID:      id,
		PondID:  pondID,
		store:   store,
		bus:     bus,
		driver:  driver,
		status:  &model.DeviceStatus{ID: id, Kind: model.DeviceAerator, PondID: pondID},
		enabled: true,
		tracker: &statTracker{},
	}
}

func (a *Aerator) Status() model.DeviceStatus {
	return *a.status
}

func (a *Aerator) Stats() Stats {
	return a.tracker.snapshot()
}

func (a *Aerator) AdjustDO(reading float64) error {
	patch := water.PatchDO(reading)
	if err := a.store.ApplyMetrics(a.PondID, patch); err != nil {
		return err
	}
	a.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, a.PondID, "aerator adjust", reading))
	return nil
}

func (a *Aerator) Start(ctx context.Context) error {
	if !a.enabled {
		return model.ErrUnavailable
	}
	if err := a.driver.Start(ctx); err != nil {
		a.handleStartFailure(err)
		return err
	}
	a.status.Succeed()
	a.tracker.onStart(time.Now())
	a.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, a.PondID, "aerator start", 1))
	return nil
}

func (a *Aerator) handleStartFailure(err error) {
	a.status.Fail(err)
	a.retryAt = time.Now().Add(30 * time.Second)
	a.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, a.PondID, "aerator failure", 1))
	a.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, a.PondID, "aerator retry scheduled", 0))
	_ = a.bus.ReportFailure(a.PondID, "aerator", err)
}

func (a *Aerator) Stop(ctx context.Context) error {
	if err := a.driver.Stop(ctx); err != nil {
		a.status.Fail(err)
		return err
	}
	a.status.Running = false
	a.status.Touch()
	a.tracker.onStop(time.Now())
	a.store.AddEvent(model.NewTelemetry(model.TelemetrySensor, a.PondID, "aerator stop", 0))
	return nil
}
