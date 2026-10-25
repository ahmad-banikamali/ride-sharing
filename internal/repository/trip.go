package repository

import (
	"context"
	"database/sql"
	"errors"
	"ride-sharing/internal/domain"
)

type TripRepository struct {
	db *sql.DB
}

func (r *TripRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	expectedStatus domain.TripStatus,
	newStatus domain.TripStatus,
) (domain.Trip, error) {
	query := `
		UPDATE trips
		SET status = $1
		WHERE id = $2 AND status = $3
		RETURNING id, pickup, destination, status
	`

	var trip domain.Trip

	err := r.db.QueryRowContext(
		ctx,
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
		return domain.Trip{}, err
	}

	_, getErr := r.GetById(ctx, id)
	if errors.Is(getErr, domain.ErrTripNotFound) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if getErr != nil {
		return domain.Trip{}, getErr
	}

	return domain.Trip{}, domain.ErrTripStatusConflict
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{
		db: db,
	}
}

func (r *TripRepository) Save(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	q := "INSERT INTO trips (pickup, destination, status) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, q, trip.Pickup, trip.Destination, trip.Status).Scan(&trip.ID)
	if err != nil {
		return domain.Trip{}, err
	}
	return trip, nil
}

func (r *TripRepository) GetById(ctx context.Context, id int64) (domain.Trip, error) {
	var trip domain.Trip
	q := "select id, pickup, destination, status from trips where id = $1"
	err := r.db.QueryRowContext(ctx, q, id).Scan(&trip.ID, &trip.Pickup, &trip.Destination, &trip.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Trip{}, domain.ErrTripNotFound
		}
		return domain.Trip{}, err
	}
	return trip, nil
}
