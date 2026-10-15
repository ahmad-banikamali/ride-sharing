package main

import (
	"fmt"
	"log"
	"net/http"
	"ride-sharing/internal/handler"
	"ride-sharing/internal/repository"
	"ride-sharing/internal/service"
)

func newRouter() *http.ServeMux {
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

	re := repository.NewTripRepository()
	ts := service.NewTripService(re)
	th := handler.NewTripHandler(ts)

	mux.HandleFunc("/trips", th.CreateTrip)
	mux.HandleFunc("GET /trips/{id}", th.GetTrip)

	return mux
}

func main() {

	mux := newRouter()

	log.Printf("Listening on port 8080\n")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
