package model

import "time"

type TelemetryKind string

const (
	TelemetrySensor TelemetryKind = "sensor"
	TelemetryAlarm  TelemetryKind = "alarm"
	TelemetryFeed   TelemetryKind = "feed"
	TelemetryPump   TelemetryKind = "pump"
	TelemetryTemp   TelemetryKind = "temp"
)

type TelemetryEvent struct {
	Kind      TelemetryKind
	PondID    string
	Detail    string
	Value     float64
	Timestamp time.Time
}

func NewTelemetry(kind TelemetryKind, pondID string, detail string, value float64) TelemetryEvent {
	return TelemetryEvent{
		Kind:      kind,
		PondID:    pondID,
		Detail:    detail,
		Value:     value,
		Timestamp: time.Now(),
	}
}
