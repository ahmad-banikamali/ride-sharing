package service

import (
	"errors"
	"ride-sharing/internal/domain"
	tripService "ride-sharing/internal/service/trip"
	"strconv"
	"testing"
)

func TestTripService_CreateTrip(t *testing.T) {
	t.Helper()
	fakeRepo := NewFakeTripRepository()
	service := tripService.NewTripService(fakeRepo)
	trip, err := service.CreateTrip(
		t.Context(),
		tripService.CreateTripInput{
			Pickup:      "A",
			Destination: "B",
		})
	if err != nil {
		t.Fatal(err)
	}
	if trip.Pickup != "A" || trip.Destination != "B" {
		t.Error("pickup or destination required")
	}

	if trip.Status != domain.TripStatusRequested {
		t.Errorf("expected status %q, got %q",
			domain.TripStatusRequested, trip.Status)
	}
}

func TestTripService_CreateTrip_EmptyData(t *testing.T) {
	t.Helper()
	fakeRepo := NewFakeTripRepository()
	service := tripService.NewTripService(fakeRepo)
	_, err := service.CreateTrip(
		t.Context(),
		tripService.CreateTripInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCanTransitionTripStatus(t *testing.T) {
	tests := []struct {
		name string
		from domain.TripStatus
		to   domain.TripStatus
		want bool
	}{
		{
			name: "requested to cancelled",
			from: domain.TripStatusRequested,
			to:   domain.TripStatusCancelled,
			want: true,
		},
		{
			name: "accepted to in progress",
			from: domain.TripStatusAccepted,
			to:   domain.TripStatusInProgress,
			want: true,
		},
		{
			name: "accepted to completed is invalid",
			from: domain.TripStatusAccepted,
			to:   domain.TripStatusCompleted,
			want: false,
		},
		{
			name: "in progress to completed",
			from: domain.TripStatusInProgress,
			to:   domain.TripStatusCompleted,
			want: true,
		},
		{
			name: "completed to in progress is invalid",
			from: domain.TripStatusCompleted,
			to:   domain.TripStatusInProgress,
			want: false,
		},
		{
			name: "cancelled to accepted is invalid",
			from: domain.TripStatusCancelled,
			to:   domain.TripStatusAccepted,
			want: false,
		},
		{
			name: "requested to requested is invalid",
			from: domain.TripStatusRequested,
			to:   domain.TripStatusRequested,
			want: false,
		},
		{
			name: "completed to completed is invalid",
			from: domain.TripStatusCompleted,
			to:   domain.TripStatusCompleted,
			want: false,
		},
		{
			name: "unknown status cannot transition",
			from: domain.TripStatus("unknown"),
			to:   domain.TripStatusAccepted,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tripService.CanTransitionTripStatus(tt.from, tt.to)

			if got != tt.want {
				t.Errorf(
					"canTransitionTripStatus(%q, %q) = %v, want %v",
					tt.from,
					tt.to,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestTripService_UpdateTripStatus_InvalidTransition(t *testing.T) {
	repo := NewFakeTripRepository()
	svc := tripService.NewTripService(repo)
	trip, err := svc.CreateTrip(
		t.Context(),
		tripService.CreateTripInput{
			Pickup:      "A",
			Destination: "B",
		})
	if err != nil {
		t.Fatal(err)
	}
	storedTrip := repo.trips[trip.ID]
	storedTrip.Status = domain.TripStatusCompleted
	repo.trips[trip.ID] = storedTrip

	input := tripService.UpdateTripStatusInput{
		ID:     strconv.FormatInt(trip.ID, 10),
		Status: string(domain.TripStatusInProgress),
	}

	_, err = svc.UpdateTripStatus(
		t.Context(),
		input)

	if !errors.Is(err, domain.ErrInvalidTripTransition) {
		t.Error("expected error")
	}

}
