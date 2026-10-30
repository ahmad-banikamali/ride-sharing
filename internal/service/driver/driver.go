package driver

import (
	"context"
	"ride-sharing/internal/domain"
	"ride-sharing/internal/service"
)

type Repository interface {
	Create(ctx context.Context, driver domain.Driver) (domain.Driver, error)
}

type Validator interface {
	ValidateInput(i *CreateInput) (bool, error)
}

type CreateInput struct {
	Name string
}

type CreateOutput struct {
	Id   int64
	Name string
}

type Service struct {
	repository Repository
	validator  InputValidator
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	validateInput, err := s.validator.ValidateInput(&input)
	ctx, cancel := context.WithTimeout(ctx, service.DbTimeout)
	defer cancel()
	if err != nil {
		return CreateOutput{}, domain.ErrDriverValidationFailed
	}

	if !validateInput {
		return CreateOutput{}, domain.ErrDriverInvalid
	}

	driver, err := s.repository.Create(ctx, domain.Driver{
		Name: input.Name,
	})

	if err != nil {
		return CreateOutput{}, domain.ErrDriverInvalid
	}
	return CreateOutput{
		Id:   driver.Id,
		Name: driver.Name,
	}, nil

}
