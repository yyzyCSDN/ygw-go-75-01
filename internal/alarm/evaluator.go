package alarm

import "aquarecirc/internal/model"

type SensorEvaluator struct{}

func NewSensorEvaluator() *SensorEvaluator {
	return &SensorEvaluator{}
}

func (e *SensorEvaluator) Evaluate(reading *model.SensorReading) (model.Alarm, bool) {
	if reading.Metric == model.MetricDO && reading.Value < 3 {
		return model.NewAlarm(reading.PondID, reading.Metric, model.LevelCritical, "low dissolved oxygen"), true
	}
	if reading.Metric == model.MetricAmmonia && reading.Value > 1.5 {
		return model.NewAlarm(reading.PondID, reading.Metric, model.LevelCritical, "high ammonia"), true
	}
	return model.Alarm{}, false
}
