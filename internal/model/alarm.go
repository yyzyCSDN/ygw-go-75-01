package model

import "time"

type Alarm struct {
	ID        string
	PondID    string
	Metric    Metric
	Level     Level
	Message   string
	Count     int
	Timestamp time.Time
}

func NewAlarm(pondID string, metric Metric, level Level, message string) Alarm {
	return Alarm{
		PondID:    pondID,
		Metric:    metric,
		Level:     level,
		Message:   message,
		Count:     1,
		Timestamp: time.Now(),
	}
}

func (a Alarm) Key() string {
	return a.PondID + "|" + string(a.Metric) + "|" + a.Level.String()
}
