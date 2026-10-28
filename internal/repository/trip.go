package repository

import (
	"context"
	"database/sql"
	"errors"
	"ride-sharing/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
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

func (r *TripRepository) AssignDriver(
	ctx context.Context,
	tripID int64,
	driverID int64,
) (domain.Trip, error) {
	query := `
        UPDATE trips
        SET driver_id = $1,
            status = 'accepted'
        WHERE id = $2
          AND status = 'requested'
          AND driver_id IS NULL
        RETURNING id, pickup, destination, status, driver_id
    `

	var trip domain.Trip
	var assignedDriverID sql.NullInt64

	err := r.db.QueryRowContext(
		ctx,
		query,
		driverID,
		tripID,
	).Scan(
		&trip.ID,
		&trip.Pickup,
		&trip.Destination,
		&trip.Status,
		&assignedDriverID,
	)

	if err == nil {
		if !assignedDriverID.Valid {
			return domain.Trip{}, domain.ErrUnknown
		}

		trip.DriverID = &assignedDriverID.Int64
		return trip, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		_, getErr := r.GetById(ctx, tripID)

		if errors.Is(getErr, domain.ErrTripNotFound) {
			return domain.Trip{}, domain.ErrTripNotFound
		}
		if getErr != nil {
			return domain.Trip{}, getErr
		}

		return domain.Trip{}, domain.ErrTripStatusConflict
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23503" &&
			pgErr.ConstraintName == "trips_driver_id_fkey":
			return domain.Trip{}, domain.ErrDriverNotFound

		case pgErr.Code == "23505" &&
			pgErr.ConstraintName == "unique_active_trip_per_driver":
			return domain.Trip{}, domain.ErrDriverUnavailable
		}
	}

	return domain.Trip{}, err
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
