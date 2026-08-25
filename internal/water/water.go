package water

import (
	"sync"
	"time"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

type MetricPatch struct {
	DO      *float64
	Ammonia *float64
	Nitrite *float64
	PH      *float64
	Temp    *float64
	Source  string
}

func PatchDO(value float64) MetricPatch {
	return MetricPatch{DO: &value, Source: "aerator"}
}

func PatchAmmonia(value float64) MetricPatch {
	return MetricPatch{Ammonia: &value, Source: "filter"}
}

type Store struct {
	mu        sync.RWMutex
	ponds     map[string]*model.Pond
	states    map[string]*model.WaterState
	balances  map[string]float64
	histories map[string]*MetricHistory
	sensors   map[string]*Sensor
	monitor   *Monitor
	bus       *alarm.Bus
	cache     *alarm.SnapshotCache
	evaluator *alarm.SensorEvaluator
	events    []model.TelemetryEvent
}

func NewStore(bus *alarm.Bus) *Store {
	return &Store{
		ponds:     map[string]*model.Pond{},
		states:    map[string]*model.WaterState{},
		balances:  map[string]float64{},
		histories: map[string]*MetricHistory{},
		sensors:   map[string]*Sensor{},
		monitor:   NewMonitor(),
		bus:       bus,
		cache:     alarm.NewSnapshotCache(),
		evaluator: alarm.NewSensorEvaluator(),
	}
}

func (s *Store) RegisterPond(p *model.Pond) error {
	if err := p.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ponds[p.ID]; ok {
		return model.ErrDuplicate
	}
	s.ponds[p.ID] = p
	s.states[p.ID] = model.NewWaterState(p.ID)
	s.balances[p.ID] = 100
	s.histories[p.ID] = newMetricHistory(200)
	return nil
}

func (s *Store) Pond(id string) (*model.Pond, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.ponds[id]
	return p, ok
}

func (s *Store) Ponds() []*model.Pond {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Pond, 0, len(s.ponds))
	for _, p := range s.ponds {
		out = append(out, p)
	}
	return out
}

func (s *Store) State(pondID string) (model.WaterState, bool) {
	s.mu.RLock()
	st, ok := s.states[pondID]
	s.mu.RUnlock()
	if !ok {
		return model.WaterState{}, false
	}
	return st.Copy(), true
}

func (s *Store) ApplyMetrics(pondID string, patch MetricPatch) error {
	st, ok := s.states[pondID]
	if !ok {
		return model.ErrNotFound
	}
	replaceMetrics(st, patch)
	st.UpdatedAt = time.Now()
	st.Level = classify(st)
	s.RecordHistory(pondID, patch, st.UpdatedAt)
	s.record(model.NewTelemetry(model.TelemetrySensor, pondID, "metrics", 0))
	return nil
}

func (s *Store) record(evt model.TelemetryEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, evt)
	if len(s.events) > 500 {
		s.events = s.events[len(s.events)-500:]
	}
}

func (s *Store) AddEvent(evt model.TelemetryEvent) {
	s.record(evt)
}

func (s *Store) MonitorOpen() int {
	return s.monitor.OpenCount()
}

func (s *Store) pollSensor(sensorID string) error {
	conn, err := s.monitor.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = s.monitor.Close(conn)
	}()
	reading, err := s.monitor.Read(conn, s, sensorID)
	if err != nil {
		return err
	}
	if reading != nil && !reading.Present {
		s.record(model.NewTelemetry(model.TelemetrySensor, sensorID, "offline", 0))
	}
	return s.ApplySensor(reading)
}

func (s *Store) PollSensor(sensorID string) error {
	return s.pollSensor(sensorID)
}

func (s *Store) CloseIdleConns(maxAge time.Duration) int {
	return s.monitor.CloseIdle(maxAge)
}


func (s *Store) ReadSensor(sensorID string) (*model.SensorReading, error) {
	s.mu.RLock()
	sensor, ok := s.sensors[sensorID]
	s.mu.RUnlock()
	if !ok {
		return nil, model.ErrNotFound
	}
	if !sensor.Online {
		return offlineReading(sensor), nil
	}
	reading := model.SensorReading{
		SensorID:  sensor.ID,
		PondID:    sensor.PondID,
		Metric:    sensor.Metric,
		Value:     sensor.Last,
		Timestamp: time.Now(),
		Present:   true,
	}
	return &reading, nil
}

func offlineReading(sensor *Sensor) *model.SensorReading {
	reading := model.EmptyReading(sensor.ID, sensor.PondID, sensor.Metric)
	reading.Timestamp = time.Now()
	return &reading
}

func (s *Store) Events(limit int) []model.TelemetryEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}
	out := make([]model.TelemetryEvent, limit)
	copy(out, s.events[len(s.events)-limit:])
	return out
}

func (s *Store) FeedBalance(pondID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balances[pondID]
}

func (s *Store) DeductBalance(pondID string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.balances[pondID]
	if !ok {
		return model.ErrNotFound
	}
	next := current - amount
	if next < 0 {
		next = 0
	}
	s.balances[pondID] = next
	return nil
}
