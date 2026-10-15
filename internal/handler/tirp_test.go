package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ride-sharing/internal/repository"
	"ride-sharing/internal/service"
	"strings"
	"testing"
)

func newTestHandler() *TripHandler {
	repo := repository.NewTripRepository()
	svc := service.NewTripService(repo)
	return NewTripHandler(svc)
}

func TestTripHandler_CreateTrip(t *testing.T) {

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			body:       `{"pickup":"Berlin","destination":"Hamburg"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       `{"pickup":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong HTTP method",
			method:     http.MethodGet,
			body:       `{"pickup":"Berlin","destination":"Hamburg"}`,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "empty pickup",
			method:     http.MethodPost,
			body:       `{"pickup":" ","destination":"Hamburg"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := newTestHandler()
			req := httptest.NewRequest(tt.method, "/api/trips", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			th.CreateTrip(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateTrip() got = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.name == "valid request" {
				var body map[string]any

				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("invalid response JSON: %v", err)
				}

				if body["pickup"] != "Berlin" {
					t.Errorf("expected pickup Berlin, got %v", body["pickup"])
				}

				if body["destination"] != "Hamburg" {
					t.Errorf("expected destination Hamburg, got %v", body["destination"])
				}
			}

		})
	}
}

func TestGetTrip(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		createTrip bool
		wantStatus int
	}{
		{
			name:       "existing trip",
			id:         "1",
			createTrip: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid trip ID",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "trip not found",
			id:         "999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()

			if tt.createTrip {
				createReq := httptest.NewRequest(
					http.MethodPost,
					"/trips",
					strings.NewReader(
						`{"pickup":"Berlin","destination":"Hamburg"}`,
					),
				)
				createRec := httptest.NewRecorder()

				handler.CreateTrip(createRec, createReq)

				if createRec.Code != http.StatusCreated {
					t.Fatalf(
						"failed to create test trip: status %d",
						createRec.Code,
					)
				}
			}

			req := httptest.NewRequest(
				http.MethodGet,
				"/trips/"+tt.id,
				nil,
			)
			req.SetPathValue("id", tt.id)

			rec := httptest.NewRecorder()
			handler.GetTrip(rec, req)

			if tt.name == "existing trip" {
				var body struct {
					ID          int    `json:"id"`
					Pickup      string `json:"pickup"`
					Destination string `json:"destination"`
				}

				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("invalid response JSON: %v", err)
				}

				if body.ID != 1 {
					t.Errorf("expected ID 1, got %d", body.ID)
				}

				if body.Pickup != "Berlin" {
					t.Errorf("expected pickup Berlin, got %q", body.Pickup)
				}

				if body.Destination != "Hamburg" {
					t.Errorf("expected destination Hamburg, got %q", body.Destination)
				}
			}

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"expected status %d, got %d; body: %s",
					tt.wantStatus,
					rec.Code,
					rec.Body.String(),
				)
			}
		})
	}
}
