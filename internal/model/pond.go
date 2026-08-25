package model

type Pond struct {
	ID          string
	Name        string
	Volume      float64
	Stock       int
	TargetDO    float64
	TargetTemp  float64
	FeedPerFish float64
	Enabled     bool
}

func NewPond(id string, name string, volume float64, stock int) *Pond {
	return &Pond{
		ID:          id,
		Name:        name,
		Volume:      volume,
		Stock:       stock,
		TargetDO:    6.0,
		TargetTemp:  26.0,
		FeedPerFish: 0.02,
		Enabled:     true,
	}
}

func (p *Pond) Validate() error {
	if p.ID == "" || p.Name == "" {
		return ErrInvalid
	}
	if p.Volume <= 0 {
		return ErrInvalid
	}
	if p.Stock < 0 {
		return ErrInvalid
	}
	if p.TargetDO <= 0 || p.TargetTemp <= 0 {
		return ErrInvalid
	}
	return nil
}

func (p *Pond) FeedAmount() float64 {
	if p.Stock <= 0 {
		return 0
	}
	return float64(p.Stock) * p.FeedPerFish
}
