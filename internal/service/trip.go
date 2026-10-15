package service

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidTripInput = errors.New("pickup or destination required")
var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTripID = errors.New("invalid trip ID")

type CreateTripInput struct {
	Pickup      string
	Destination string
}

type GetTripInput struct {
	ID string
}

type Trip struct {
	ID          int
	Pickup      string
	Destination string
}

type TripRepository interface {
	Save(trip Trip) (Trip, error)
	GetById(id int) (Trip, error)
}

type TripService struct {
	tripRepository TripRepository
}

func NewTripService(repository TripRepository) *TripService {
	return &TripService{
		tripRepository: repository,
	}
}

func (s *TripService) CreateTrip(input CreateTripInput) (Trip, error) {
	//validate input
	input.Pickup = strings.TrimSpace(input.Pickup)
	input.Destination = strings.TrimSpace(input.Destination)

	if input.Pickup == "" || input.Destination == "" {
		return Trip{}, ErrInvalidTripInput
	}

	trip, err := s.tripRepository.Save(Trip{
		Pickup:      input.Pickup,
		Destination: input.Destination,
	})
	if err != nil {
		return Trip{}, err
	}

	return trip, nil
}

func (s *TripService) GetById(input GetTripInput) (Trip, error) {
	id, err := strconv.Atoi(input.ID)
	if err != nil {
		return Trip{}, ErrInvalidTripID
	}
	trip, err := s.tripRepository.GetById(id)
	if err != nil {
		return Trip{}, err
	}
	return trip, nil
}
