package pump

import (
	"context"
	"errors"
	"testing"

	"aquarecirc/internal/alarm"
	"aquarecirc/internal/model"
	"aquarecirc/internal/water"
)

func TestPumpRestartTimeoutNotSwallowed(t *testing.T) {
	bus := alarm.NewBus()
	store := water.NewStore(bus)
	_ = store.RegisterPond(model.NewPond("p1", "A", 100, 1000))
	driver := NewSimDriver()
	driver.SetStartError(errors.New("motor stuck"))
	unit := New("pmp-1", "p1", store, driver, bus)
	err := unit.Restart(context.Background())
	if err == nil {
		t.Fatal("pump restart error was swallowed")
	}
	st := unit.Status()
	if st.State == model.PumpRunning {
		t.Fatal("pump marked running although restart failed")
	}
}
