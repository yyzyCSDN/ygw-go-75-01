package water

import (
	"sync"
	"time"

	"aquarecirc/internal/model"
)

type HistoryPoint struct {
	Metric model.Metric
	Value  float64
	At     time.Time
}

type MetricHistory struct {
	mu     sync.Mutex
	points map[model.Metric][]HistoryPoint
	limit  int
}

func newMetricHistory(limit int) *MetricHistory {
	return &MetricHistory{
		points: map[model.Metric][]HistoryPoint{},
		limit:  limit,
	}
}

func (h *MetricHistory) Add(metric model.Metric, value float64, at time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.points[metric]
	bucket = append(bucket, HistoryPoint{Metric: metric, Value: value, At: at})
	if len(bucket) > h.limit {
		bucket = bucket[len(bucket)-h.limit:]
	}
	h.points[metric] = bucket
}

func (h *MetricHistory) Recent(metric model.Metric, limit int) []HistoryPoint {
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.points[metric]
	if limit <= 0 || limit > len(bucket) {
		limit = len(bucket)
	}
	out := make([]HistoryPoint, limit)
	copy(out, bucket[len(bucket)-limit:])
	return out
}

func (s *Store) RecordHistory(pondID string, patch MetricPatch, at time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.recordHistoryLocked(pondID, patch, at)
}

// recordHistoryLocked records the patched metrics into the pond's history.
// The caller must hold s.mu (at least RLock); each history.MetricHistory guards
// its own slice with its own mutex, so this stays safe under the read lock.
func (s *Store) recordHistoryLocked(pondID string, patch MetricPatch, at time.Time) {
	history := s.histories[pondID]
	if history == nil {
		return
	}
	if patch.DO != nil {
		history.Add(model.MetricDO, *patch.DO, at)
	}
	if patch.Ammonia != nil {
		history.Add(model.MetricAmmonia, *patch.Ammonia, at)
	}
	if patch.Nitrite != nil {
		history.Add(model.MetricNitrite, *patch.Nitrite, at)
	}
	if patch.PH != nil {
		history.Add(model.MetricPH, *patch.PH, at)
	}
	if patch.Temp != nil {
		history.Add(model.MetricTemp, *patch.Temp, at)
	}
}

func (s *Store) History(pondID string, metric model.Metric, limit int) []HistoryPoint {
	s.mu.RLock()
	history := s.histories[pondID]
	s.mu.RUnlock()
	if history == nil {
		return nil
	}
	return history.Recent(metric, limit)
}
