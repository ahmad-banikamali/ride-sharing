package driver

type InputValidator struct {
}

func (v InputValidator) ValidateInput(i *CreateInput) (bool, error) {
	//todo add logical validation
	return true, nil
}
