package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"ride-sharing/internal/domain"
	"ride-sharing/internal/service"
)

type CreateTripRequest struct {
	Pickup      string `json:"pickup"`
	Destination string `json:"destination"`
}

type UpdateTripStatusRequest struct {
	Status string `json:"status"`
}

type TripResponse struct {
	ID          int64  `json:"id"`
	Pickup      string `json:"pickup"`
	Destination string `json:"destination"`
	Status      string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type TripHandler struct {
	service *service.TripService
}

func NewTripHandler(service *service.TripService) *TripHandler {
	return &TripHandler{service: service}
}

func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	trip, err := h.service.CreateTrip(
		r.Context(),
		service.CreateTripInput{
			Pickup:      req.Pickup,
			Destination: req.Destination,
		})
	if err != nil {
		writeServiceError(w, "create trip", err)
		return
	}

	writeJSON(w, http.StatusCreated, toTripResponse(trip))
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	trip, err := h.service.GetById(
		r.Context(),
		service.GetTripInput{
			ID: r.PathValue("id"),
		})
	if err != nil {
		writeServiceError(w, "get trip", err)
		return
	}

	writeJSON(w, http.StatusOK, toTripResponse(trip))
}

func (h *TripHandler) UpdateTripStatus(w http.ResponseWriter, r *http.Request) {
	var req UpdateTripStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	trip, err := h.service.UpdateTripStatus(
		r.Context(),
		service.UpdateTripStatusInput{
			ID:     r.PathValue("id"),
			Status: req.Status,
		})
	if err != nil {
		writeServiceError(w, "update trip status", err)
		return
	}

	writeJSON(w, http.StatusOK, toTripResponse(trip))
}

func toTripResponse(trip domain.Trip) TripResponse {
	return TripResponse{
		ID:          trip.ID,
		Pickup:      trip.Pickup,
		Destination: trip.Destination,
		Status:      string(trip.Status),
	}
}

// writeServiceError is the single place where domain errors become HTTP status
// codes. Unmapped errors are infrastructure failures: their detail stays in the
// log and never reaches the client.
func writeServiceError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTripInput):
		writeError(w, http.StatusBadRequest, "pickup or destination required")
	case errors.Is(err, domain.ErrInvalidTripID):
		writeError(w, http.StatusBadRequest, "invalid trip ID")
	case errors.Is(err, domain.ErrTripNotFound):
		writeError(w, http.StatusNotFound, "trip not found")
	case errors.Is(err, domain.ErrInvalidTripStatus):
		writeError(w, http.StatusUnprocessableEntity, "invalid trip status")
	case errors.Is(err, domain.ErrInvalidTripTransition):
		writeError(w, http.StatusConflict, "invalid trip status transition")
	case errors.Is(err, domain.ErrTripStatusConflict):
		writeError(w, http.StatusConflict, "trip status changed concurrently")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line is already on the wire; logging is all that is left.
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
