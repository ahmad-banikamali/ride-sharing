package service

type FakeTripRepository struct {
	trips  map[int64]Trip
	nextID int64
}

func NewFakeTripRepository() *FakeTripRepository {
	return &FakeTripRepository{
		trips:  make(map[int64]Trip),
		nextID: 1,
	}
}

func (r *FakeTripRepository) Save(trip Trip) (Trip, error) {
	trip.ID = r.nextID
	r.nextID++

	r.trips[trip.ID] = trip
	return trip, nil
}

func (r *FakeTripRepository) GetById(id int64) (Trip, error) {
	trip, ok := r.trips[id]
	if !ok {
		return Trip{}, ErrTripNotFound
	}

	return trip, nil
}

func NewFakeService() *TripService {
	fakeRepo := NewFakeTripRepository()
	service := NewTripService(fakeRepo)
	return service
}
