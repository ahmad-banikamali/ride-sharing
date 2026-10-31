package service

import (
	"context"
	"fmt"
	"ride-sharing/internal/driver"
	"time"
)

const dbTimeout = 5 * time.Second

type Repository interface {
	Create(ctx context.Context, d driver.Driver) (driver.Driver, error)
}

type Validator interface {
	Validate(name string) error
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

func NewService(
	repository Repository,
	validator Validator) *Service {
	return &Service{
		repository: repository,
		validator:  validator,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	err := s.validator.Validate(input.Name)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	if err != nil {
		return CreateOutput{}, driver.ErrInvalidDriverInput
	}

	created, err := s.repository.Create(ctx, driver.Driver{
		Name: input.Name,
	})

	if err != nil {
		return CreateOutput{}, fmt.Errorf("create driver error: %w", err)
	}
	return CreateOutput{
		Id:   created.ID,
		Name: created.Name,
	}, nil

}
