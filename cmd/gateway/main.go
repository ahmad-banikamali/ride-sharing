package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"ride-sharing/internal/database"
	"ride-sharing/internal/handler"
	"ride-sharing/internal/repository"
	"ride-sharing/internal/service"
)

func newRouter(th *handler.TripHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Printf("failed to write health check response: %v", err)
			return
		}
	})

	mux.HandleFunc("/trips", th.CreateTrip)
	mux.HandleFunc("GET /trips/{id}", th.GetTrip)

	return mux
}

func main() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:local_dev_password@localhost:5432/ride_sharing"
	}

	db, err := database.NewPostgres(dsn)
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
