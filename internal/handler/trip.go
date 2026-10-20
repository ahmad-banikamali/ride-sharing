package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"ride-sharing/internal/service"
)

type CreateTripRequest struct {
	Pickup      string `json:"pickup"`
	Destination string `json:"destination"`
}

type TripResponse struct {
	ID          int64  `json:"id"`
	Pickup      string `json:"pickup"`
	Destination string `json:"destination"`
	Status      string `json:"status"`
}

type TripHandler struct {
	service *service.TripService
}

func NewTripHandler(service *service.TripService) *TripHandler {
	return &TripHandler{service: service}
}

func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	input := service.CreateTripInput{
		Pickup:      req.Pickup,
		Destination: req.Destination,
	}
	trip, err := h.service.CreateTrip(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTripInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	res := TripResponse{
		ID:          trip.ID,
		Pickup:      trip.Pickup,
		Destination: trip.Destination,
		Status:      string(trip.Status),
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		// در این نقطه ممکن است پاسخ HTTP قبلاً ارسال شده باشد.
		// بنابراین نمی‌توانیم با اطمینان وضعیت پاسخ را تغییر دهیم.
		return
	}
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//id := r.URL.Query().Get("id")
	id := r.PathValue("id") // /trips/12
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	input := service.GetTripInput{
		ID: id,
	}
	byId, err := h.service.GetById(input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTripID):
			{
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		case errors.Is(err, service.ErrTripNotFound):
			{
				http.Error(w, err.Error(), http.StatusNotFound)
			}

		default:
			{
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		return
	}

	res := TripResponse{
		ID:          byId.ID,
		Pickup:      byId.Pickup,
		Destination: byId.Destination,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
