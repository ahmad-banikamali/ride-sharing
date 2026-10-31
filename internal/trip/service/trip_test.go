package service

import (
	"errors"
	"ride-sharing/internal/trip"
	"strconv"
	"testing"
)

func TestTripService_CreateTrip(t *testing.T) {
	t.Helper()
	fakeRepo := NewFakeTripRepository()
	service := NewTripService(fakeRepo)
	created, err := service.CreateTrip(
		t.Context(),
		CreateTripInput{
			Pickup:      "A",
			Destination: "B",
		})
	if err != nil {
		t.Fatal(err)
	}
	if created.Pickup != "A" || created.Destination != "B" {
		t.Error("pickup or destination required")
	}

	if created.Status != trip.StatusRequested {
		t.Errorf("expected status %q, got %q",
			trip.StatusRequested, created.Status)
	}
}

func TestTripService_CreateTrip_EmptyData(t *testing.T) {
	t.Helper()
	fakeRepo := NewFakeTripRepository()
	service := NewTripService(fakeRepo)
	_, err := service.CreateTrip(
		t.Context(),
		CreateTripInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCanTransitionTripStatus(t *testing.T) {
	tests := []struct {
		name string
		from trip.Status
		to   trip.Status
		want bool
	}{
		{
			name: "requested to cancelled",
			from: trip.StatusRequested,
			to:   trip.StatusCancelled,
			want: true,
		},
		{
			name: "accepted to in progress",
			from: trip.StatusAccepted,
			to:   trip.StatusInProgress,
			want: true,
		},
		{
			name: "accepted to completed is invalid",
			from: trip.StatusAccepted,
			to:   trip.StatusCompleted,
			want: false,
		},
		{
			name: "in progress to completed",
			from: trip.StatusInProgress,
			to:   trip.StatusCompleted,
			want: true,
		},
		{
			name: "completed to in progress is invalid",
			from: trip.StatusCompleted,
			to:   trip.StatusInProgress,
			want: false,
		},
		{
			name: "cancelled to accepted is invalid",
			from: trip.StatusCancelled,
			to:   trip.StatusAccepted,
			want: false,
		},
		{
			name: "requested to requested is invalid",
			from: trip.StatusRequested,
			to:   trip.StatusRequested,
			want: false,
		},
		{
			name: "completed to completed is invalid",
			from: trip.StatusCompleted,
			to:   trip.StatusCompleted,
			want: false,
		},
		{
			name: "unknown status cannot transition",
			from: trip.Status("unknown"),
			to:   trip.StatusAccepted,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanTransitionTripStatus(tt.from, tt.to)

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
	svc := NewTripService(repo)
	created, err := svc.CreateTrip(
		t.Context(),
		CreateTripInput{
			Pickup:      "A",
			Destination: "B",
		})
	if err != nil {
		t.Fatal(err)
	}
	storedTrip := repo.trips[created.ID]
	storedTrip.Status = trip.StatusCompleted
	repo.trips[created.ID] = storedTrip

	input := UpdateTripStatusInput{
		ID:     strconv.FormatInt(created.ID, 10),
		Status: string(trip.StatusInProgress),
	}

	_, err = svc.UpdateTripStatus(
		t.Context(),
		input)

	if !errors.Is(err, trip.ErrInvalidTripTransition) {
		t.Error("expected error")
	}

}
