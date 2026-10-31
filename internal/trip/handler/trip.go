package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"ride-sharing/internal/trip"
	"ride-sharing/internal/trip/service"
	"strconv"
)

type CreateTripRequest struct {
	Pickup      string `json:"pickup"`
	Destination string `json:"destination"`
}

type UpdateTripStatusRequest struct {
	Status string `json:"status"`
}

type AssignDriverRequest struct {
	DriverID string `json:"driver_id"`
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

func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{service: svc}
}

func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	created, err := h.service.CreateTrip(
		r.Context(),
		service.CreateTripInput{
			Pickup:      req.Pickup,
			Destination: req.Destination,
		})
	if err != nil {
		writeServiceError(w, "create trip", err)
		return
	}

	writeJSON(w, http.StatusCreated, toTripResponse(created))
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.GetById(
		r.Context(),
		service.GetTripInput{
			ID: r.PathValue("id"),
		})
	if err != nil {
		writeServiceError(w, "get trip", err)
		return
	}

	writeJSON(w, http.StatusOK, toTripResponse(found))
}

func (h *TripHandler) UpdateTripStatus(w http.ResponseWriter, r *http.Request) {
	var req UpdateTripStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	updated, err := h.service.UpdateTripStatus(
		r.Context(),
		service.UpdateTripStatusInput{
			ID:     r.PathValue("id"),
			Status: req.Status,
		})
	if err != nil {
		writeServiceError(w, "update trip status", err)
		return
	}

	writeJSON(w, http.StatusOK, toTripResponse(updated))
}

func (h *TripHandler) AssignDriver(w http.ResponseWriter, r *http.Request) {
	tripId := r.PathValue("id")
	var input AssignDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
	}
	driverId, err := strconv.ParseInt(input.DriverID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid driver ID")
	}

	driver, err := h.service.AssignDriver(r.Context(), service.AssignDriverInput{
		TripID:   tripId,
		DriverID: driverId,
	})

	if err != nil {
		writeServiceError(w, "assign driver", err)
	}

	writeJSON(w, http.StatusOK, toTripResponse(driver))
}

func toTripResponse(t trip.Trip) TripResponse {
	return TripResponse{
		ID:          t.ID,
		Pickup:      t.Pickup,
		Destination: t.Destination,
		Status:      string(t.Status),
	}
}

// writeServiceError is the single place where domain errors become HTTP status
// codes. Unmapped errors are infrastructure failures: their detail stays in the
// log and never reaches the client.
func writeServiceError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, trip.ErrInvalidTripInput):
		writeError(w, http.StatusBadRequest, "pickup or destination required")
	case errors.Is(err, trip.ErrInvalidTripID):
		writeError(w, http.StatusBadRequest, "invalid trip ID")
	case errors.Is(err, trip.ErrTripNotFound):
		writeError(w, http.StatusNotFound, "trip not found")
	case errors.Is(err, trip.ErrInvalidTripStatus):
		writeError(w, http.StatusUnprocessableEntity, "invalid trip status")
	case errors.Is(err, trip.ErrInvalidTripTransition):
		writeError(w, http.StatusConflict, "invalid trip status transition")
	case errors.Is(err, trip.ErrTripStatusConflict):
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
