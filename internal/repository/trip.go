package repository

import (
	"ride-sharing/internal/service"
	"sync"
)

type TripRepository struct {
	trips  []service.Trip
	mu     sync.RWMutex
	nextID int
}

func NewTripRepository() *TripRepository {
	return &TripRepository{
		trips:  make([]service.Trip, 0),
		nextID: 1,
	}
}

func (r *TripRepository) Save(trip service.Trip) (service.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	trip.ID = r.nextID
	r.trips = append(r.trips, trip)
	r.nextID++
	return trip, nil
}

func (r *TripRepository) GetById(id int) (service.Trip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, trip := range r.trips {
		if trip.ID == id {
			return trip, nil
		}
	}

	return service.Trip{}, service.ErrTripNotFound

}
