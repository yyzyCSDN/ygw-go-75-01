package model

import "time"

type DeviceKind string

const (
	DeviceAerator DeviceKind = "aerator"
	DeviceFilter  DeviceKind = "filter"
	DevicePump    DeviceKind = "pump"
	DeviceHeater  DeviceKind = "heater"
)

type DeviceStatus struct {
	ID        string
	Kind      DeviceKind
	PondID    string
	Running   bool
	LastError string
	LastSeen  time.Time
}

func (d *DeviceStatus) Touch() {
	d.LastSeen = time.Now()
}

func (d *DeviceStatus) Fail(err error) {
	d.Running = false
	if err != nil {
		d.LastError = err.Error()
	}
	d.LastSeen = time.Now()
}

func (d *DeviceStatus) Succeed() {
	d.Running = true
	d.LastError = ""
	d.LastSeen = time.Now()
}
