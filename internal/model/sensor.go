package model

import "time"

type Metric string

const (
	MetricDO      Metric = "do"
	MetricAmmonia Metric = "ammonia"
	MetricNitrite Metric = "nitrite"
	MetricPH      Metric = "ph"
	MetricTemp    Metric = "temp"
	MetricDevice  Metric = "device"
)

func ParseMetric(value string) (Metric, bool) {
	switch Metric(value) {
	case MetricDO, MetricAmmonia, MetricNitrite, MetricPH, MetricTemp:
		return Metric(value), true
	}
	return "", false
}

type SensorReading struct {
	SensorID  string
	PondID    string
	Metric    Metric
	Value     float64
	Timestamp time.Time
	Present   bool
}

func EmptyReading(sensorID string, pondID string, metric Metric) SensorReading {
	return SensorReading{
		SensorID: sensorID,
		PondID:   pondID,
		Metric:   metric,
		Present:  false,
	}
}

func (r *SensorReading) Summary() string {
	if r == nil {
		return "empty"
	}
	if !r.Present {
		return "offline"
	}
	return string(r.Metric)
}
