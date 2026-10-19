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
	})
	if err != nil {
		t.Fatal(err)
	}
	if save.ID <= 0 {
		t.Errorf("ID was not set %v", save.ID)
	}
	t.Cleanup(func() {
		_, err := repo.db.Exec("DELETE FROM trips WHERE id = 1")
		if err != nil {
			return
		}
	})
}

func TestTripRepository_GetByID(t *testing.T) {
	t.Helper()
	repo, cleanup := newTestRepository(t)
	t.Cleanup(cleanup)
	t.Cleanup(func() {
		_, err := repo.db.Exec("DELETE FROM trips WHERE id = 1")
		if err != nil {
			t.Fatal(err)
		}
	})

	save, err := repo.Save(service.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
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
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := newTestRepository(t)
	t.Cleanup(cleanup)
	_, err := repo.GetById(1)
	if !errors.Is(err, service.ErrTripNotFound) {
		t.Error("expected error")
	}
}
