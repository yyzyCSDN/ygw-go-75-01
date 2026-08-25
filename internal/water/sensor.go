package water

import "aquarecirc/internal/model"

type Sensor struct {
	ID     string
	PondID string
	Metric model.Metric
	Last   float64
	Online bool
}
