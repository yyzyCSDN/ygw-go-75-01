package pump

import (
	"sync"
	"time"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

type Pump struct {
	ID     string
	PondID string
	mu     sync.Mutex
	status model.PumpStatus
	device model.DeviceStatus
	store  *water.Store
	driver Driver
	bus    *alarm.Bus
	rotor  *rotationTracker
}

func New(id string, pondID string, store *water.Store, driver Driver, bus *alarm.Bus) *Pump {
	return &Pump{
		ID:     id,
		PondID: pondID,
		status: model.NewPumpStatus(pondID),
		device: model.DeviceStatus{ID: id, Kind: model.DevicePump, PondID: pondID},
		store:  store,
		driver: driver,
		bus:    bus,
		rotor:  newRotationTracker(120),
	}
}

func (p *Pump) Status() model.PumpStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

func (p *Pump) BeginFeeding() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status.State == model.PumpRestarting {
		return model.ErrBusy
	}
	p.status.State = model.PumpStopped
	p.status.ActualRPM = 0
	p.device.Running = false
	p.device.Touch()
	return nil
}

func (p *Pump) EndFeeding() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status.State == model.PumpRestarting {
		return model.ErrBusy
	}
	p.status.State = model.PumpRunning
	p.status.ActualRPM = p.status.TargetRPM
	p.device.Running = true
	p.device.Touch()
	return nil
}

func (p *Pump) DeviceStatus() model.DeviceStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.device
}

func (p *Pump) RecordRotation(rpm int) {
	p.rotor.add(rpm, time.Now())
}

func (p *Pump) Health() Health {
	status := p.Status()
	return p.rotor.health(status.TargetRPM)
}
