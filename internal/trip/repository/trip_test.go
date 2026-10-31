package repository

import (
	"context"
	"errors"
	"math"
	"os"
	"ride-sharing/internal/platform/database"
	"ride-sharing/internal/trip"
	"testing"
)

const defaultTestDSN = "postgres://app:local_dev_password@localhost:5432/ride_sharing"

// newTestRepository opens a real PostgreSQL connection and registers cleanup.
// Skipped under `go test -short`, so DB-free runs stay possible.
func newTestRepository(t *testing.T) *TripRepository {
	t.Helper()

	if testing.Short() {
		t.Skip("integration test: needs PostgreSQL, skipped with -short")
	}

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	db, err := database.NewPostgres(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return &TripRepository{db: db}
}

// createTrip inserts a requested trip and registers its deletion.
func createTrip(t *testing.T, repo *TripRepository) trip.Trip {
	t.Helper()

	created, err := repo.Save(context.Background(), trip.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
		Status:      trip.StatusRequested,
	})
	if err != nil {
		t.Fatalf("save trip: %v", err)
	}

	t.Cleanup(func() {
		if _, err := repo.db.Exec("DELETE FROM trips WHERE id = $1", created.ID); err != nil {
			t.Errorf("delete test trip %d: %v", created.ID, err)
		}
	})

	return created
}

func TestTripRepository_Save(t *testing.T) {
	repo := newTestRepository(t)
	created := createTrip(t, repo)

	if created.ID <= 0 {
		t.Errorf("ID was not set: %v", created.ID)
	}

	if created.Status != trip.StatusRequested {
		t.Errorf("expected status %q, got %q", trip.StatusRequested, created.Status)
	}
}

func TestTripRepository_GetByID(t *testing.T) {
	repo := newTestRepository(t)
	created := createTrip(t, repo)

	got, err := repo.GetById(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get trip: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Pickup != created.Pickup {
		t.Errorf("expected pickup %q, got %q", created.Pickup, got.Pickup)
	}
	if got.Destination != created.Destination {
		t.Errorf("expected destination %q, got %q", created.Destination, got.Destination)
	}
	if got.Status != trip.StatusRequested {
		t.Errorf("expected status %q, got %q", trip.StatusRequested, got.Status)
	}
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepository(t)

	// math.MaxInt64 can never be produced by the identity sequence.
	_, err := repo.GetById(context.Background(), math.MaxInt64)
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound, got %v", err)
	}
}

func TestTripRepository_UpdateStatus(t *testing.T) {
	repo := newTestRepository(t)
	created := createTrip(t, repo)

	updated, err := repo.UpdateStatus(
		context.Background(),
		created.ID,
		trip.StatusRequested,
		trip.StatusCancelled,
	)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != trip.StatusCancelled {
		t.Errorf("expected status %q, got %q", trip.StatusCancelled, updated.Status)
	}

	// RETURNING shows what the statement produced; re-reading proves what was stored.
	stored, err := repo.GetById(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get trip: %v", err)
	}
	if stored.Status != trip.StatusCancelled {
		t.Errorf("expected persisted status %q, got %q", trip.StatusCancelled, stored.Status)
	}
}

func TestTripRepository_UpdateStatus_NotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.UpdateStatus(
		context.Background(),
		math.MaxInt64,
		trip.StatusRequested,
		trip.StatusCancelled,
	)
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound, got %v", err)
	}
}

func TestTripRepository_UpdateStatus_Conflict(t *testing.T) {
	repo := newTestRepository(t)
	created := createTrip(t, repo)

	// First transition wins: requested -> canceled.
	if _, err := repo.UpdateStatus(
		context.Background(),
		created.ID,
		trip.StatusRequested,
		trip.StatusCancelled,
	); err != nil {
		t.Fatalf("first update: %v", err)
	}

	// Second call sends a stale expectation: the row exists,
	// but the WHERE predicate no longer matches.
	_, err := repo.UpdateStatus(
		context.Background(),
		created.ID,
		trip.StatusRequested,
		trip.StatusCancelled,
	)
	if !errors.Is(err, trip.ErrTripStatusConflict) {
		t.Fatalf("expected ErrTripStatusConflict, got %v", err)
	}
}
