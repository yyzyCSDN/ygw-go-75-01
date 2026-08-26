package pump

import (
	"context"

	"aquarecirc/internal/model"
)

func (p *Pump) Restart(ctx context.Context) error {
	p.mu.Lock()
	p.status.State = model.PumpRestarting
	p.mu.Unlock()
	_ = p.driver.Start(context.Background())
	p.mu.Lock()
	p.status.State = model.PumpRunning
	p.status.ActualRPM = p.status.TargetRPM
	p.mu.Unlock()
	p.device.Succeed()
	p.store.AddEvent(model.NewTelemetry(model.TelemetryPump, p.PondID, "pump restart", 1))
	return nil
}
