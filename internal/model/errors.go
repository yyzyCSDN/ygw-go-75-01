package model

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrDuplicate   = errors.New("already exists")
	ErrInvalid     = errors.New("invalid argument")
	ErrTimeout     = errors.New("operation timed out")
	ErrClosed      = errors.New("already closed")
	ErrBusy        = errors.New("component busy")
	ErrUnavailable = errors.New("component unavailable")
)
