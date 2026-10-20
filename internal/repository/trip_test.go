package repository

import (
	"errors"
	"ride-sharing/internal/database"
	"ride-sharing/internal/service"
	"testing"
)

func newTestRepository(t *testing.T) (*TripRepository, func()) {
	t.Helper()
	db, err := database.NewPostgres("postgres://app:local_dev_password@localhost:5432/ride_sharing")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return &TripRepository{db: db}, cleanup
}

func TestTripRepository_Save(t *testing.T) {
	repo, cleanup := newTestRepository(t)
	t.Cleanup(cleanup)
	save, err := repo.Save(service.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
		Status:      service.TripStatusRequested,
	})
	if err != nil {
		t.Fatal(err)
	}

	if save.ID <= 0 {
		t.Errorf("ID was not set: %v", save.ID)
	}

	if save.Status != service.TripStatusRequested {
		t.Errorf("expected status %q, got %q",
			service.TripStatusRequested, save.Status)
	}

	t.Cleanup(func() {
		_, err := repo.db.Exec("DELETE FROM trips WHERE id = $1", save.ID)
		if err != nil {
			t.Errorf("failed to delete test trip: %v", err)
		}
	})
}
func TestTripRepository_GetByID(t *testing.T) {
	t.Helper()
	repo, cleanup := newTestRepository(t)
	t.Cleanup(cleanup)
	save, err := repo.Save(service.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
		Status:      service.TripStatusRequested,
	})
	t.Cleanup(func() {
		_, err := repo.db.Exec("DELETE FROM trips WHERE id = $1", save.ID)
		if err != nil {
			t.Fatal(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	id, err := repo.GetById(save.ID)
	if err != nil {
		t.Fatal(err)
	}
	if id.Pickup != "Tehran" {
		t.Errorf("Pickup was not set %v", id)
	}
	if id.Destination != "Mashhad" {
		t.Errorf("Destination was not set %v", id)
	}
	if id.Status != service.TripStatusRequested {
		t.Errorf("expected status %q, got %q",
			service.TripStatusRequested, id.Status)
	}
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := newTestRepository(t)
	t.Cleanup(cleanup)
	_, err := repo.GetById(1)
	if !errors.Is(err, service.ErrTripNotFound) {
		t.Error("expected error")
	}
}
