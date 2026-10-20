package service

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidTripInput = errors.New("pickup or destination required")
var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTripID = errors.New("invalid trip ID")

var ErrInvalidTripStatus = errors.New("invalid trip status")
var ErrInvalidTripTransition = errors.New("invalid trip status transition")
var ErrTripStatusConflict = errors.New("trip status changed concurrently")

type CreateTripInput struct {
	Pickup      string
	Destination string
}

type GetTripInput struct {
	ID string
}

type UpdateTripStatusInput struct {
	ID     string
	Status string
}

type TripStatus string

const (
	TripStatusRequested  TripStatus = "requested"
	TripStatusAccepted   TripStatus = "accepted"
	TripStatusInProgress TripStatus = "in_progress"
	TripStatusCompleted  TripStatus = "completed"
	TripStatusCancelled  TripStatus = "cancelled"
)

func isValidTripStatus(status TripStatus) bool {
	switch status {
	case
		TripStatusRequested,
		TripStatusAccepted,
		TripStatusInProgress,
		TripStatusCompleted,
		TripStatusCancelled:
		return true
	default:
		return false
	}
}

var validTripStatusTransitions = map[TripStatus][]TripStatus{
	TripStatusRequested: {
		TripStatusAccepted,
		TripStatusCancelled,
	},
	TripStatusAccepted: {
		TripStatusInProgress,
		TripStatusCancelled,
	},
	TripStatusInProgress: {
		TripStatusCompleted,
	},
}

type Trip struct {
	ID          int64
	Pickup      string
	Destination string
	Status      TripStatus
}

type TripRepository interface {
	Save(trip Trip) (Trip, error)
	GetById(id int64) (Trip, error)
	UpdateStatus(
		id int64,
		expectedStatus TripStatus,
		newStatus TripStatus,
	) (Trip, error)
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
		Status:      TripStatusRequested,
	})
	if err != nil {
		return Trip{}, err
	}

	return trip, nil
}

func (s *TripService) GetById(input GetTripInput) (Trip, error) {
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return Trip{}, ErrInvalidTripID
	}
	trip, err := s.tripRepository.GetById(id)
	if err != nil {
		return Trip{}, err
	}
	return trip, nil
}

func (s *TripService) UpdateStatus(input UpdateTripStatusInput) (Trip, error) {
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return Trip{}, ErrInvalidTripID
	}
	newStatus := TripStatus(input.Status)
	if !isValidTripStatus(newStatus) {
		return Trip{}, ErrInvalidTripStatus
	}
	trip, err := s.tripRepository.GetById(id)
	if err != nil {
		return Trip{}, ErrTripNotFound
	}
	currentStatus := trip.Status

	canTransit := canTransitionTripStatus(currentStatus, newStatus)
	if canTransit {
		trip.Status = newStatus
	}
	return s.tripRepository.UpdateStatus(id, currentStatus, newStatus)
}

func canTransitionTripStatus(from, to TripStatus) bool {

	for _, validTo := range validTripStatusTransitions[from] {
		if to == validTo {
			return true
		}
	}

	return false
}
