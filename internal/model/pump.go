package model

type PumpState int

const (
	PumpStopped PumpState = iota
	PumpRunning
	PumpRestarting
)

type PumpStatus struct {
	PondID    string
	State     PumpState
	TargetRPM int
	ActualRPM int
	LastError string
}

func NewPumpStatus(pondID string) PumpStatus {
	return PumpStatus{
		PondID:    pondID,
		State:     PumpStopped,
		TargetRPM: 1200,
	}
}
