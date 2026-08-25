package water

import (
	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

func (s *Store) RegisterSensor(sensor *Sensor) error {
	if sensor.ID == "" {
		return model.ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sensors[sensor.ID] = sensor
	return nil
}

func (s *Store) Sensors() []*Sensor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Sensor, 0, len(s.sensors))
	for _, sensor := range s.sensors {
		out = append(out, sensor)
	}
	return out
}

func (s *Store) ApplySensor(reading *model.SensorReading) error {
	if reading.SensorID == "" || reading.PondID == "" {
		return model.ErrInvalid
	}
	if !reading.Present {
		return nil
	}
	patch := MetricPatch{Source: reading.SensorID}
	value := reading.Value
	switch reading.Metric {
	case model.MetricDO:
		patch.DO = &value
	case model.MetricAmmonia:
		patch.Ammonia = &value
	case model.MetricNitrite:
		patch.Nitrite = &value
	case model.MetricPH:
		patch.PH = &value
	case model.MetricTemp:
		patch.Temp = &value
	default:
		return model.ErrInvalid
	}
	if err := s.ApplyMetrics(reading.PondID, patch); err != nil {
		return err
	}
	if evt, ok := s.evaluator.Evaluate(reading); ok {
		key := alarm.NewReportKey(reading.PondID, string(reading.Metric), evt.Level.String())
		if s.bus.Acquire(key) {
			_ = s.bus.Report(evt)
		}
	}
	s.reportLevel(reading.PondID)
	return nil
}

func (s *Store) reportLevel(pondID string) {
	st, ok := s.State(pondID)
	if !ok {
		return
	}
	if st.Level.Severity() < model.LevelWarning.Severity() {
		return
	}
	key := alarm.NewReportKey(pondID, "water", st.Level.String())
	if !s.bus.Acquire(key) {
		return
	}
	evt := model.NewAlarm(pondID, model.MetricDO, st.Level, "water level "+st.Level.String())
	_ = s.bus.Report(evt)
}
