package driver

import "errors"

type Driver struct {
	ID   int64
	Name string
}

var ErrDriverInvalid = errors.New("invalid driver input")
var ErrInvalidDriverInput = errors.New("validation failed")
