package repository

import (
	"ride-sharing/internal/service"
	"testing"
)

func TestTripRepository_Save(t *testing.T) {
	repo := NewTripRepository()
	save, err := repo.Save(service.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
	})
	if err != nil {
		t.Fatal(err)
	}
	if save.ID != 1 {
		t.Errorf("ID was not set %v", save.ID)
	}
}

func TestTripRepository_GetByID(t *testing.T) {
	repo := NewTripRepository()
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
	repo := NewTripRepository()
	_, err := repo.GetById(1)
	if err == nil {
		t.Error("expected error")
	}
}
