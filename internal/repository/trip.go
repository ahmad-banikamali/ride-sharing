package repository

import (
	"database/sql"
	"errors"
	"ride-sharing/internal/service"
)

type TripRepository struct {
	db *sql.DB
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{
		db: db,
	}
}

func (r *TripRepository) Save(trip service.Trip) (service.Trip, error) {
	q := "INSERT INTO trips (pickup, destination) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRow(q, trip.Pickup, trip.Destination).Scan(&trip.ID)
	if err != nil {
		return service.Trip{}, err
	}
	return trip, nil
}

func (r *TripRepository) GetById(id int64) (service.Trip, error) {
	var trip service.Trip
	q := "select id, pickup, destination from trips where id = $1"
	err := r.db.QueryRow(q, id).Scan(&trip.ID, &trip.Pickup, &trip.Destination)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.Trip{}, service.ErrTripNotFound
		}
		return service.Trip{}, err
	}
	return trip, nil

}
