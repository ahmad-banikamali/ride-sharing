package driver

import (
	"ride-sharing/internal/service/driver"
)

type InputValidator struct {
}

func (v InputValidator) Validate(i *driver.CreateInput) error {
	//todo add logical validation
	return nil
}
