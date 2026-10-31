package repository

import (
	"context"
	"database/sql"
	"errors"
	"ride-sharing/internal/trip"

	"github.com/jackc/pgx/v5/pgconn"
)

type TripRepository struct {
	db *sql.DB
}

func (r *TripRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	expectedStatus trip.Status,
	newStatus trip.Status,
) (trip.Trip, error) {
	query := `
		UPDATE trips
		SET status = $1
		WHERE id = $2 AND status = $3
		RETURNING id, pickup, destination, status
	`

	var result trip.Trip

	err := r.db.QueryRowContext(
		ctx,
		query,
		newStatus,
		id,
		expectedStatus,
	).Scan(
		&result.ID,
		&result.Pickup,
		&result.Destination,
		&result.Status,
	)

	if err == nil {
		return result, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return trip.Trip{}, err
	}

	_, getErr := r.GetById(ctx, id)
	if errors.Is(getErr, trip.ErrTripNotFound) {
		return trip.Trip{}, trip.ErrTripNotFound
	}
	if getErr != nil {
		return trip.Trip{}, getErr
	}

	return trip.Trip{}, trip.ErrTripStatusConflict
}

func (r *TripRepository) AssignDriver(
	ctx context.Context,
	tripID int64,
	driverID int64,
) (trip.Trip, error) {
	query := `
        UPDATE trips
        SET driver_id = $1,
            status = 'accepted'
        WHERE id = $2
          AND status = 'requested'
          AND driver_id IS NULL
        RETURNING id, pickup, destination, status, driver_id
    `

	var result trip.Trip
	var assignedDriverID sql.NullInt64

	err := r.db.QueryRowContext(
		ctx,
		query,
		driverID,
		tripID,
	).Scan(
		&result.ID,
		&result.Pickup,
		&result.Destination,
		&result.Status,
		&assignedDriverID,
	)

	if err == nil {
		if !assignedDriverID.Valid {
			return trip.Trip{}, trip.ErrUnknown
		}

		result.DriverID = &assignedDriverID.Int64
		return result, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		_, getErr := r.GetById(ctx, tripID)

		if errors.Is(getErr, trip.ErrTripNotFound) {
			return trip.Trip{}, trip.ErrTripNotFound
		}
		if getErr != nil {
			return trip.Trip{}, getErr
		}

		return trip.Trip{}, trip.ErrTripStatusConflict
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23503" &&
			pgErr.ConstraintName == "trips_driver_id_fkey":
			return trip.Trip{}, trip.ErrDriverNotFound

		case pgErr.Code == "23505" &&
			pgErr.ConstraintName == "unique_active_trip_per_driver":
			return trip.Trip{}, trip.ErrDriverUnavailable
		}
	}

	return trip.Trip{}, err
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{
		db: db,
	}
}

func (r *TripRepository) Save(ctx context.Context, t trip.Trip) (trip.Trip, error) {
	q := "INSERT INTO trips (pickup, destination, status) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, q, t.Pickup, t.Destination, t.Status).Scan(&t.ID)
	if err != nil {
		return trip.Trip{}, err
	}
	return t, nil
}

func (r *TripRepository) GetById(ctx context.Context, id int64) (trip.Trip, error) {
	var result trip.Trip
	q := "select id, pickup, destination, status from trips where id = $1"
	err := r.db.QueryRowContext(ctx, q, id).Scan(&result.ID, &result.Pickup, &result.Destination, &result.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return trip.Trip{}, trip.ErrTripNotFound
		}
		return trip.Trip{}, err
	}
	return result, nil
}
