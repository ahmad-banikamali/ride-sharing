package service

import "testing"

type fakeTripRepository struct{}

func (fakeTripRepository) Save(trip Trip) (Trip, error) {
	return trip, nil
}

func (fakeTripRepository) GetById(id int) (Trip, error) {
	return Trip{}, nil
}

func TestTripService_CreateTrip(t *testing.T) {
	repository := &fakeTripRepository{}
	service := NewTripService(repository)
	trip, err := service.CreateTrip(CreateTripInput{
		Pickup:      "A",
		Destination: "B",
	})
	if err != nil {
		t.Fatal(err)
	}
	if trip.Pickup != "A" || trip.Destination != "B" {
		t.Error("pickup or destination required")
	}
}

func TestTripService_CreateTrip_EmptyData(t *testing.T) {
	repository := &fakeTripRepository{}
	service := NewTripService(repository)
	_, err := service.CreateTrip(CreateTripInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}
