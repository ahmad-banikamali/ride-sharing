package repository

import (
	"database/sql"
	"errors"
	"ride-sharing/internal/service"
)

type TripRepository struct {
	db *sql.DB
}

func (r *TripRepository) UpdateStatus(
	id int64,
	expectedStatus service.TripStatus,
	newStatus service.TripStatus,
) (service.Trip, error) {
	query := `
		UPDATE trips
		SET status = $1
		WHERE id = $2 AND status = $3
		RETURNING id, pickup, destination, status
	`

	var trip service.Trip

	err := r.db.QueryRow(
		query,
		newStatus,
		id,
		expectedStatus,
	).Scan(
		&trip.ID,
		&trip.Pickup,
		&trip.Destination,
		&trip.Status,
	)

	if err == nil {
		return trip, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return service.Trip{}, err
	}

	_, getErr := r.GetById(id)
	if errors.Is(getErr, service.ErrTripNotFound) {
		return service.Trip{}, service.ErrTripNotFound
	}
	if getErr != nil {
		return service.Trip{}, getErr
	}

	return service.Trip{}, service.ErrTripStatusConflict
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{
		db: db,
	}
}

func (r *TripRepository) Save(trip service.Trip) (service.Trip, error) {
	q := "INSERT INTO trips (pickup, destination, status) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(q, trip.Pickup, trip.Destination, trip.Status).Scan(&trip.ID)
	if err != nil {
		return service.Trip{}, err
	}
	return trip, nil
}

func (r *TripRepository) GetById(id int64) (service.Trip, error) {
	var trip service.Trip
	q := "select id, pickup, destination, status from trips where id = $1"
	err := r.db.QueryRow(q, id).Scan(&trip.ID, &trip.Pickup, &trip.Destination, &trip.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.Trip{}, service.ErrTripNotFound
		}
		return service.Trip{}, err
	}
	return trip, nil

}
