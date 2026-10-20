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

	if trip.Status != TripStatusRequested {
		t.Errorf("expected status %q, got %q",
			TripStatusRequested, trip.Status)
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

func TestCanTransitionTripStatus(t *testing.T) {
	tests := []struct {
		name string
		from TripStatus
		to   TripStatus
		want bool
	}{
		{
			name: "requested to accepted",
			from: TripStatusRequested,
			to:   TripStatusAccepted,
			want: true,
		},
		{
			name: "requested to cancelled",
			from: TripStatusRequested,
			to:   TripStatusCancelled,
			want: true,
		},
		{
			name: "accepted to in progress",
			from: TripStatusAccepted,
			to:   TripStatusInProgress,
			want: true,
		},
		{
			name: "accepted to completed is invalid",
			from: TripStatusAccepted,
			to:   TripStatusCompleted,
			want: false,
		},
		{
			name: "in progress to completed",
			from: TripStatusInProgress,
			to:   TripStatusCompleted,
			want: true,
		},
		{
			name: "completed to in progress is invalid",
			from: TripStatusCompleted,
			to:   TripStatusInProgress,
			want: false,
		},
		{
			name: "cancelled to accepted is invalid",
			from: TripStatusCancelled,
			to:   TripStatusAccepted,
			want: false,
		},
		{
			name: "requested to requested is invalid",
			from: TripStatusRequested,
			to:   TripStatusRequested,
			want: false,
		},
		{
			name: "completed to completed is invalid",
			from: TripStatusCompleted,
			to:   TripStatusCompleted,
			want: false,
		},
		{
			name: "unknown status cannot transition",
			from: TripStatus("unknown"),
			to:   TripStatusAccepted,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canTransitionTripStatus(tt.from, tt.to)

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
