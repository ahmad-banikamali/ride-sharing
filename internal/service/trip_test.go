package service

import (
	"testing"
)

func TestTripService_CreateTrip(t *testing.T) {
	t.Helper()
	service := NewFakeService()
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
	t.Helper()
	service := NewFakeService()
	_, err := service.CreateTrip(CreateTripInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}
