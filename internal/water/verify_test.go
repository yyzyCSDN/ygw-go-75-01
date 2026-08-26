package water

import (
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

func TestMonitorConnClosed(t *testing.T) {
	bus := alarm.NewBus()
	store := NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	_ = store.RegisterSensor(&Sensor{ID: "p1-do", PondID: "p1", Metric: model.MetricDO, Last: 6.0, Online: true})
	if err := store.PollSensor("p1-do"); err != nil {
		t.Fatalf("poll: %v", err)
	}
	if open := store.MonitorOpen(); open != 0 {
		t.Fatalf("monitor connection leaked: %d open", open)
	}
}
