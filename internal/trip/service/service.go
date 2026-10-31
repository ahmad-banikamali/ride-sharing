package service

import (
	"context"
	"fmt"
	"ride-sharing/internal/trip"
	"slices"
	"strconv"
	"strings"
	"time"
)

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

type AssignDriverInput struct {
	TripID   string
	DriverID int64
}

const dbTimeout = 5 * time.Second

var validTripStatusTransitions = map[trip.Status][]trip.Status{
	trip.StatusRequested: {
		// trip.StatusAccepted, // only assign method can change requested status to accepted status
		trip.StatusCancelled,
	},
	trip.StatusAccepted: {
		trip.StatusInProgress,
		trip.StatusCancelled,
	},
	trip.StatusInProgress: {
		trip.StatusCompleted,
	},
}

type TripRepository interface {
	Save(ctx context.Context, t trip.Trip) (trip.Trip, error)
	GetById(ctx context.Context, id int64) (trip.Trip, error)
	UpdateStatus(ctx context.Context,
		id int64,
		expectedStatus trip.Status,
		newStatus trip.Status,
	) (trip.Trip, error)

	AssignDriver(
		ctx context.Context,
		tripID int64,
		driverID int64,
	) (trip.Trip, error)
}

type TripService struct {
	tripRepository TripRepository
}

func NewTripService(repository TripRepository) *TripService {
	return &TripService{
		tripRepository: repository,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, input CreateTripInput) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	//validate input
	input.Pickup = strings.TrimSpace(input.Pickup)
	input.Destination = strings.TrimSpace(input.Destination)

	if input.Pickup == "" || input.Destination == "" {
		return trip.Trip{}, trip.ErrInvalidTripInput
	}
	created, err := s.tripRepository.Save(ctx, trip.Trip{
		Pickup:      input.Pickup,
		Destination: input.Destination,
		Status:      trip.StatusRequested,
	})
	if err != nil {
		return trip.Trip{}, fmt.Errorf("save trip: %w", err)
	}

	return created, nil
}

func (s *TripService) GetById(ctx context.Context, input GetTripInput) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return trip.Trip{}, trip.ErrInvalidTripID
	}
	found, err := s.tripRepository.GetById(ctx, id)
	if err != nil {
		return trip.Trip{}, err
	}
	return found, nil
}

func (s *TripService) UpdateTripStatus(ctx context.Context, input UpdateTripStatusInput) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return trip.Trip{}, trip.ErrInvalidTripID
	}

	newStatus := trip.Status(input.Status)
	if !isValidTripStatus(newStatus) {
		return trip.Trip{}, trip.ErrInvalidTripStatus
	}

	current, err := s.tripRepository.GetById(ctx, id)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("get trip %d: %w", id, err)
	}

	if !CanTransitionTripStatus(current.Status, newStatus) {
		return trip.Trip{}, trip.ErrInvalidTripTransition
	}

	// The conditional UPDATE re-checks the status, so a change that happens
	// between the read above and this write cannot slip through.
	updated, err := s.tripRepository.UpdateStatus(ctx, id, current.Status, newStatus)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("update trip %d status: %w", id, err)
	}

	return updated, nil
}

func (s *TripService) AssignDriver(ctx context.Context, input AssignDriverInput) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.TripID, 10, 64)
	if err != nil {
		return trip.Trip{}, trip.ErrInvalidTripID
	}
	driver, err := s.tripRepository.AssignDriver(ctx, id, input.DriverID)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("assign driver %d: %w", id, err)
	}
	return driver, nil
}

func isValidTripStatus(status trip.Status) bool {
	switch status {
	case
		trip.StatusRequested,
		trip.StatusAccepted,
		trip.StatusInProgress,
		trip.StatusCompleted,
		trip.StatusCancelled:
		return true
	default:
		return false
	}
}

func CanTransitionTripStatus(from, to trip.Status) bool {
	return slices.Contains(validTripStatusTransitions[from], to)
}
