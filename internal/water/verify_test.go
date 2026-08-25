package water

import (
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
)

func TestEmptySensorNoNilPanic(t *testing.T) {
	bus := alarm.NewBus()
	store := NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	if err := store.RegisterSensor(&Sensor{ID: "s-off", PondID: "p1", Metric: model.MetricTemp, Online: false}); err != nil {
		t.Fatalf("register sensor: %v", err)
	}
	reading, err := store.ReadSensor("s-off")
	if err != nil {
		t.Fatalf("read sensor: %v", err)
	}
	if reading == nil {
		t.Fatal("empty sensor returned nil reading")
	}
	if reading.Present {
		t.Fatal("offline sensor must not report present data")
	}
	if err := store.ApplySensor(reading); err != nil {
		t.Fatalf("apply sensor: %v", err)
	}
}
