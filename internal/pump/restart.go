package pump

import (
	"context"
	"time"

	"aquarecirc/internal/model"
)

func (p *Pump) Restart(ctx context.Context) error {
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	p.mu.Lock()
	p.status.State = model.PumpRestarting
	p.mu.Unlock()
	if err := p.driver.Start(timeout); err != nil {
		p.failRestart(err)
		return err
	}
	p.mu.Lock()
	p.status.State = model.PumpRunning
	p.status.ActualRPM = p.status.TargetRPM
	p.mu.Unlock()
	p.device.Succeed()
	p.store.AddEvent(model.NewTelemetry(model.TelemetryPump, p.PondID, "pump restart", 1))
	return nil
}

func (p *Pump) failRestart(err error) {
	p.mu.Lock()
	p.status.State = model.PumpStopped
	p.status.LastError = err.Error()
	p.status.ActualRPM = 0
	p.mu.Unlock()
	p.device.Fail(err)
	_ = p.bus.ReportFailure(p.PondID, "pump", err)
	p.rotor.add(0, time.Now())
	p.store.AddEvent(model.NewTelemetry(model.TelemetryPump, p.PondID, "pump restart failure", 0))
}
