package model

type TempStatus int

const (
	TempAdjusting TempStatus = iota
	TempStable
	TempAborted
)

type TempRegulation struct {
	PondID   string
	Target   float64
	Current  float64
	Status   TempStatus
	Attempts int
}

func NewTempRegulation(pondID string, target float64) TempRegulation {
	return TempRegulation{
		PondID:   pondID,
		Target:   target,
		Status:   TempAdjusting,
		Attempts: 1,
	}
}

func (r *TempRegulation) Abort() {
	r.Status = TempAborted
}

func (r *TempRegulation) Stabilize(current float64) {
	r.Current = current
	r.Status = TempStable
}
