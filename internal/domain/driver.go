package domain

import "errors"

type Driver struct {
	Id   int64
	Name string
}

var ErrDriverInvalid = errors.New("invalid driver input")
var ErrDriverValidationFailed = errors.New("validation failed")
