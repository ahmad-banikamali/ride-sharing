package service

import (
	"context"
	"ride-sharing/internal/domain"
)

type FakeTripRepository struct {
	trips  map[int64]domain.Trip
	nextID int64
}

func (r *FakeTripRepository) UpdateStatus(
	ctx context.Context,
	id int64, expectedStatus domain.TripStatus, newStatus domain.TripStatus) (domain.Trip, error) {
	current, ok := r.trips[id]
	if !ok {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if current.Status != expectedStatus {
		return domain.Trip{}, domain.ErrTripStatusConflict
	}
	current.Status = newStatus
	r.trips[id] = current
	return current, nil
}

func NewFakeTripRepository() *FakeTripRepository {
	return &FakeTripRepository{
		trips:  make(map[int64]domain.Trip),
		nextID: 1,
	}
}

func (r *FakeTripRepository) Save(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	trip.ID = r.nextID
	r.nextID++

	r.trips[trip.ID] = trip
	return trip, nil
}

func (r *FakeTripRepository) GetById(ctx context.Context, id int64) (domain.Trip, error) {
	trip, ok := r.trips[id]
	if !ok {
		return domain.Trip{}, domain.ErrTripNotFound
	}

	return trip, nil
}
