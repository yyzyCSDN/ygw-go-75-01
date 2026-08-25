package model

import "time"

type Level int

const (
	LevelNormal Level = iota
	LevelWarning
	LevelCritical
)

func (l Level) String() string {
	switch l {
	case LevelWarning:
		return "warning"
	case LevelCritical:
		return "critical"
	default:
		return "normal"
	}
}

func (l Level) Severity() int {
	return int(l)
}

type WaterState struct {
	PondID    string
	DO        float64
	Ammonia   float64
	Nitrite   float64
	PH        float64
	Temp      float64
	Level     Level
	UpdatedAt time.Time
}

func NewWaterState(pondID string) *WaterState {
	return &WaterState{
		PondID:    pondID,
		DO:        6.2,
		Ammonia:   0.2,
		Nitrite:   0.05,
		PH:        7.0,
		Temp:      26,
		Level:     LevelNormal,
		UpdatedAt: time.Now(),
	}
}

func (s *WaterState) Copy() WaterState {
	if s == nil {
		return WaterState{}
	}
	return *s
}

type WaterSnapshot struct {
	PondID  string
	DO      float64
	Ammonia float64
	Nitrite float64
	PH      float64
	Temp    float64
	Level   Level
}

func SnapshotOf(st WaterState) WaterSnapshot {
	return WaterSnapshot{
		PondID:  st.PondID,
		DO:      st.DO,
		Ammonia: st.Ammonia,
		Nitrite: st.Nitrite,
		PH:      st.PH,
		Temp:    st.Temp,
		Level:   st.Level,
	}
}
