package pump

import (
	"sync"
	"time"
)

type RotationSample struct {
	RPM int
	At  time.Time
}

type Health struct {
	SampleCount int
	MinRPM      float64
	MaxRPM      float64
	AvgRPM      float64
	Deviation   float64
	Degraded    bool
}

type rotationTracker struct {
	mu      sync.Mutex
	samples []RotationSample
	limit   int
}

func newRotationTracker(limit int) *rotationTracker {
	return &rotationTracker{
		samples: []RotationSample{},
		limit:   limit,
	}
}

func (t *rotationTracker) add(rpm int, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samples = append(t.samples, RotationSample{RPM: rpm, At: at})
	if len(t.samples) > t.limit {
		t.samples = t.samples[len(t.samples)-t.limit:]
	}
}

func (t *rotationTracker) health(target int) Health {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.samples) == 0 {
		return Health{MinRPM: float64(target), MaxRPM: float64(target), AvgRPM: float64(target)}
	}
	min := float64(t.samples[0].RPM)
	max := float64(t.samples[0].RPM)
	sum := 0.0
	for _, sample := range t.samples {
		value := float64(sample.RPM)
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
		sum += value
	}
	avg := sum / float64(len(t.samples))
	deviation := 0.0
	for _, sample := range t.samples {
		diff := float64(sample.RPM) - avg
		if diff < 0 {
			diff = -diff
		}
		if diff > deviation {
			deviation = diff
		}
	}
	return Health{
		SampleCount: len(t.samples),
		MinRPM:      min,
		MaxRPM:      max,
		AvgRPM:      avg,
		Deviation:   deviation,
		Degraded:    deviation > 120 || avg < float64(target)*0.7,
	}
}
