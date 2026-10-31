package service

import (
	"context"
	"ride-sharing/internal/trip"
)

type FakeTripRepository struct {
	trips  map[int64]trip.Trip
	nextID int64
}

func (r *FakeTripRepository) AssignDriver(
	ctx context.Context,
	tripID int64,
	driverID int64,
) (trip.Trip, error) {

	return r.trips[tripID], nil
}

func (r *FakeTripRepository) UpdateStatus(
	ctx context.Context,
	id int64, expectedStatus trip.Status, newStatus trip.Status) (trip.Trip, error) {
	current, ok := r.trips[id]
	if !ok {
		return trip.Trip{}, trip.ErrTripNotFound
	}
	if current.Status != expectedStatus {
		return trip.Trip{}, trip.ErrTripStatusConflict
	}
	current.Status = newStatus
	r.trips[id] = current
	return current, nil
}

func NewFakeTripRepository() *FakeTripRepository {
	return &FakeTripRepository{
		trips:  make(map[int64]trip.Trip),
		nextID: 1,
	}
}

func (r *FakeTripRepository) Save(ctx context.Context, t trip.Trip) (trip.Trip, error) {
	t.ID = r.nextID
	r.nextID++

	r.trips[t.ID] = t
	return t, nil
}

func (r *FakeTripRepository) GetById(ctx context.Context, id int64) (trip.Trip, error) {
	found, ok := r.trips[id]
	if !ok {
		return trip.Trip{}, trip.ErrTripNotFound
	}

	return found, nil
}
