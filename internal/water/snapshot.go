package water

import (
	"time"

	"aquarecirc/internal/model"
)

func (s *Store) TakeSnapshot(pondID string) (model.WaterSnapshot, bool) {
	st, ok := s.State(pondID)
	if !ok {
		return model.WaterSnapshot{}, false
	}
	snap := model.SnapshotOf(st)
	snap.PondID = pondID
	s.record(model.NewTelemetry(model.TelemetrySensor, pondID, "snapshot", 0))
	return snap, true
}

func (s *Store) WriteBackSnapshot(pondID string, snap model.WaterSnapshot) {
	s.cache.WriteBack(pondID, &snap)
}

func (s *Store) LatestSnapshot(pondID string) (model.WaterSnapshot, bool) {
	return s.cache.Latest(pondID)
}

func (s *Store) SnapshotCacheSize() int {
	return s.cache.Size()
}

func (s *Store) SampleAll() int {
	ponds := s.Ponds()
	count := 0
	for _, pond := range ponds {
		snap, ok := s.TakeSnapshot(pond.ID)
		if !ok {
			continue
		}
		s.WriteBackSnapshot(pond.ID, snap)
		count++
	}
	return count
}

func (s *Store) LastSampleAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.events) - 1; i >= 0; i-- {
		if s.events[i].Kind == model.TelemetrySensor && s.events[i].Detail == "snapshot" {
			return s.events[i].Timestamp
		}
	}
	return time.Time{}
}
