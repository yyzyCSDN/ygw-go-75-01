package pump

import (
	"context"
)

type Driver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type SimDriver struct {
	startErr error
}

func NewSimDriver() *SimDriver {
	return &SimDriver{}
}

func (d *SimDriver) SetStartError(err error) {
	d.startErr = err
}

func (d *SimDriver) Start(ctx context.Context) error {
	return d.startErr
}

func (d *SimDriver) Stop(ctx context.Context) error {
	return nil
}
