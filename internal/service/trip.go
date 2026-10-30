package service

import (
	"context"
	"fmt"
	"ride-sharing/internal/domain"
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

var validTripStatusTransitions = map[domain.TripStatus][]domain.TripStatus{
	domain.TripStatusRequested: {
		//domain.TripStatusAccepted, // only assign method can change requested status to accepted status
		domain.TripStatusCancelled,
	},
	domain.TripStatusAccepted: {
		domain.TripStatusInProgress,
		domain.TripStatusCancelled,
	},
	domain.TripStatusInProgress: {
		domain.TripStatusCompleted,
	},
}

type TripRepository interface {
	Save(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	GetById(ctx context.Context, id int64) (domain.Trip, error)
	UpdateStatus(ctx context.Context,
		id int64,
		expectedStatus domain.TripStatus,
		newStatus domain.TripStatus,
	) (domain.Trip, error)

	AssignDriver(
		ctx context.Context,
		tripID int64,
		driverID int64,
	) (domain.Trip, error)
}

type TripService struct {
	tripRepository TripRepository
}

func NewTripService(repository TripRepository) *TripService {
	return &TripService{
		tripRepository: repository,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, input CreateTripInput) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	//validate input
	input.Pickup = strings.TrimSpace(input.Pickup)
	input.Destination = strings.TrimSpace(input.Destination)

	if input.Pickup == "" || input.Destination == "" {
		return domain.Trip{}, domain.ErrInvalidTripInput
	}
	trip, err := s.tripRepository.Save(ctx, domain.Trip{
		Pickup:      input.Pickup,
		Destination: input.Destination,
		Status:      domain.TripStatusRequested,
	})
	if err != nil {
		return domain.Trip{}, fmt.Errorf("save trip: %w", err)
	}

	return trip, nil
}

func (s *TripService) GetById(ctx context.Context, input GetTripInput) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return domain.Trip{}, domain.ErrInvalidTripID
	}
	trip, err := s.tripRepository.GetById(ctx, id)
	if err != nil {
		return domain.Trip{}, err
	}
	return trip, nil
}

func (s *TripService) UpdateTripStatus(ctx context.Context, input UpdateTripStatusInput) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return domain.Trip{}, domain.ErrInvalidTripID
	}

	newStatus := domain.TripStatus(input.Status)
	if !isValidTripStatus(newStatus) {
		return domain.Trip{}, domain.ErrInvalidTripStatus
	}

	current, err := s.tripRepository.GetById(ctx, id)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip %d: %w", id, err)
	}

	if !canTransitionTripStatus(current.Status, newStatus) {
		return domain.Trip{}, domain.ErrInvalidTripTransition
	}

	// The conditional UPDATE re-checks the status, so a change that happens
	// between the read above and this write cannot slip through.
	trip, err := s.tripRepository.UpdateStatus(ctx, id, current.Status, newStatus)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("update trip %d status: %w", id, err)
	}

	return trip, nil
}

func (s *TripService) AssignDriver(ctx context.Context, input AssignDriverInput) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	id, err := strconv.ParseInt(input.TripID, 10, 64)
	if err != nil {
		return domain.Trip{}, domain.ErrInvalidTripID
	}
	driver, err := s.tripRepository.AssignDriver(ctx, id, input.DriverID)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("assign driver %d: %w", id, err)
	}
	return driver, nil
}

func isValidTripStatus(status domain.TripStatus) bool {
	switch status {
	case
		domain.TripStatusRequested,
		domain.TripStatusAccepted,
		domain.TripStatusInProgress,
		domain.TripStatusCompleted,
		domain.TripStatusCancelled:
		return true
	default:
		return false
	}
}

func canTransitionTripStatus(from, to domain.TripStatus) bool {

	for _, validTo := range validTripStatusTransitions[from] {
		if to == validTo {
			return true
		}
	}

	return false
}
