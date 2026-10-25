package repository

import (
	"context"
	"errors"
	"math"
	"os"
	"ride-sharing/internal/database"
	"ride-sharing/internal/domain"
	"sync"
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
func createTrip(t *testing.T, repo *TripRepository) domain.Trip {
	t.Helper()

	trip, err := repo.Save(context.Background(), domain.Trip{
		Pickup:      "Tehran",
		Destination: "Mashhad",
		Status:      domain.TripStatusRequested,
	})
	if err != nil {
		t.Fatalf("save trip: %v", err)
	}

	t.Cleanup(func() {
		if _, err := repo.db.Exec("DELETE FROM trips WHERE id = $1", trip.ID); err != nil {
			t.Errorf("delete test trip %d: %v", trip.ID, err)
		}
	})

	return trip
}

func TestTripRepository_Save(t *testing.T) {
	repo := newTestRepository(t)
	trip := createTrip(t, repo)

	if trip.ID <= 0 {
		t.Errorf("ID was not set: %v", trip.ID)
	}

	if trip.Status != domain.TripStatusRequested {
		t.Errorf("expected status %q, got %q", domain.TripStatusRequested, trip.Status)
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
	if got.Status != domain.TripStatusRequested {
		t.Errorf("expected status %q, got %q", domain.TripStatusRequested, got.Status)
	}
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepository(t)

	// math.MaxInt64 can never be produced by the identity sequence.
	_, err := repo.GetById(context.Background(), math.MaxInt64)
	if !errors.Is(err, domain.ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound, got %v", err)
	}
}

func TestTripRepository_UpdateStatus(t *testing.T) {
	repo := newTestRepository(t)
	trip := createTrip(t, repo)

	updated, err := repo.UpdateStatus(
		context.Background(),
		trip.ID,
		domain.TripStatusRequested,
		domain.TripStatusAccepted,
	)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != domain.TripStatusAccepted {
		t.Errorf("expected status %q, got %q", domain.TripStatusAccepted, updated.Status)
	}

	// RETURNING shows what the statement produced; re-reading proves what was stored.
	stored, err := repo.GetById(context.Background(), trip.ID)
	if err != nil {
		t.Fatalf("get trip: %v", err)
	}
	if stored.Status != domain.TripStatusAccepted {
		t.Errorf("expected persisted status %q, got %q", domain.TripStatusAccepted, stored.Status)
	}
}

func TestTripRepository_UpdateStatus_NotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.UpdateStatus(
		context.Background(),
		math.MaxInt64,
		domain.TripStatusRequested,
		domain.TripStatusAccepted,
	)
	if !errors.Is(err, domain.ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound, got %v", err)
	}
}

func TestTripRepository_UpdateStatus_Conflict(t *testing.T) {
	repo := newTestRepository(t)
	trip := createTrip(t, repo)

	// First transition wins: requested -> accepted.
	if _, err := repo.UpdateStatus(
		context.Background(),
		trip.ID,
		domain.TripStatusRequested,
		domain.TripStatusAccepted,
	); err != nil {
		t.Fatalf("first update: %v", err)
	}

	// Second call sends a stale expectation: the row exists,
	// but the WHERE predicate no longer matches.
	_, err := repo.UpdateStatus(
		context.Background(),
		trip.ID,
		domain.TripStatusRequested,
		domain.TripStatusCancelled,
	)
	if !errors.Is(err, domain.ErrTripStatusConflict) {
		t.Fatalf("expected ErrTripStatusConflict, got %v", err)
	}
}

func TestTripRepository_UpdateStatus_ConcurrentAccept(t *testing.T) {
	repo := newTestRepository(t)
	trip := createTrip(t, repo)

	const attempts = 10

	start := make(chan struct{})
	results := make(chan error, attempts)

	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // all goroutines fire together
			_, err := repo.UpdateStatus(
				context.Background(),
				trip.ID,
				domain.TripStatusRequested,
				domain.TripStatusAccepted,
			)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var winners, conflicts int
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, domain.ErrTripStatusConflict):
			conflicts++
		default:
			t.Fatalf("unexpected error from concurrent update: %v", err)
		}
	}

	if winners != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", winners)
	}
	if conflicts != attempts-1 {
		t.Fatalf("expected %d conflicts, got %d", attempts-1, conflicts)
	}
}
