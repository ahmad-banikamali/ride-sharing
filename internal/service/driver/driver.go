package driver

import (
	"context"
	"fmt"
	"ride-sharing/internal/domain"
	"time"
)

const dbTimeout = 5 * time.Second

type Repository interface {
	Create(ctx context.Context, driver domain.Driver) (domain.Driver, error)
}

type Validator interface {
	Validate(i *CreateInput) error
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
	validator  Validator
}

func NewService(validator Validator) *Service {
	return &Service{
		validator: validator,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	err := s.validator.Validate(&input)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	if err != nil {
		return CreateOutput{}, domain.ErrInvalidDriverInput
	}

	driver, err := s.repository.Create(ctx, domain.Driver{
		Name: input.Name,
	})

	if err != nil {
		return CreateOutput{}, fmt.Errorf("create driver error: %w", err)
	}
	return CreateOutput{
		Id:   driver.ID,
		Name: driver.Name,
	}, nil

}
