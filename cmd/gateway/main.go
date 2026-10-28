package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"ride-sharing/internal/database"
	"ride-sharing/internal/handler"
	"ride-sharing/internal/repository"
	"ride-sharing/internal/service"
	"time"
)

func newRouter(th *handler.TripHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Printf("failed to write health check response: %v", err)
		}
	})

	mux.HandleFunc("POST /trips", th.CreateTrip)
	mux.HandleFunc("GET /trips/{id}", th.GetTrip)
	mux.HandleFunc("PATCH /trips/{id}/status", th.UpdateTripStatus)
	mux.HandleFunc("POST /trips/{id}", th.AssignDriver)

	return mux
}

func main() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:local_dev_password@localhost:5432/ride_sharing"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	log.Println("Connected to PostgreSQL")

	re := repository.NewTripRepository(db)
	ts := service.NewTripService(re)
	th := handler.NewTripHandler(ts)

	mux := newRouter(th)

	log.Printf("Listening on port 8080\n")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
