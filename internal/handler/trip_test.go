package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ride-sharing/internal/domain"
	"ride-sharing/internal/service"
	"strconv"
	"strings"
	"testing"
)

func newTestHandler() *TripHandler {
	fakeRepo := service.NewFakeTripRepository()
	svc := service.NewTripService(fakeRepo)
	return NewTripHandler(svc)
}

func createTripForTest(t *testing.T, h *TripHandler) int64 {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/trips",
		strings.NewReader(`{"pickup":"Berlin","destination":"Hamburg"}`),
	)
	rec := httptest.NewRecorder()
	h.CreateTrip(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create trip: status %d", rec.Code)
	}

	var body TripResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid create response JSON: %v", err)
	}

	return body.ID
}

func patchStatus(t *testing.T, h *TripHandler, id, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/trips/"+id+"/status",
		strings.NewReader(body),
	)
	req.SetPathValue("id", id)

	rec := httptest.NewRecorder()
	h.UpdateTripStatus(rec, req)

	return rec
}

func TestTripHandler_CreateTrip(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       `{"pickup":"Berlin","destination":"Hamburg"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			body:       `{"pickup":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty pickup",
			body:       `{"pickup":" ","destination":"Hamburg"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			req := httptest.NewRequest(http.MethodPost, "/trips", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.CreateTrip(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusCreated {
				return
			}

			var body TripResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}

			if body.ID == 0 {
				t.Error("expected a generated ID")
			}
			if body.Pickup != "Berlin" || body.Destination != "Hamburg" {
				t.Errorf("unexpected payload: %+v", body)
			}
			if body.Status != string(domain.TripStatusRequested) {
				t.Errorf("expected status %q, got %q", domain.TripStatusRequested, body.Status)
			}
		})
	}
}

func TestTripHandler_GetTrip(t *testing.T) {
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
			h := newTestHandler()

			if tt.createTrip {
				createTripForTest(t, h)
			}

			req := httptest.NewRequest(http.MethodGet, "/trips/"+tt.id, nil)
			req.SetPathValue("id", tt.id)

			rec := httptest.NewRecorder()
			h.GetTrip(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d; body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var body TripResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}

			if body.ID != 1 || body.Pickup != "Berlin" || body.Destination != "Hamburg" {
				t.Errorf("unexpected payload: %+v", body)
			}
			// This locks the regression: the trip's status must be visible.
			if body.Status != string(domain.TripStatusRequested) {
				t.Errorf("expected status %q, got %q", domain.TripStatusRequested, body.Status)
			}
		})
	}
}

func TestTripHandler_UpdateTripStatus(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		createTrip bool
		bodies     []string
		wantCode   int
		wantStatus string
	}{
		{
			name:       "requested to cancelled",
			createTrip: true,
			bodies:     []string{`{"status":"cancelled"}`},
			wantCode:   http.StatusOK,
			wantStatus: string(domain.TripStatusCancelled),
		},
		{
			name:       "accepted to cancelled",
			createTrip: true,
			bodies:     []string{`{"status":"cancelled"}`},
			wantCode:   http.StatusOK,
			wantStatus: string(domain.TripStatusCancelled),
		},
		{
			name:       "requested to completed is illegal",
			createTrip: true,
			bodies:     []string{`{"status":"completed"}`},
			wantCode:   http.StatusConflict,
		},
		{
			name:       "cancelled is terminal",
			createTrip: true,
			bodies:     []string{`{"status":"cancelled"}`, `{"status":"accepted"}`},
			wantCode:   http.StatusConflict,
		},
		{
			name:       "unknown status value",
			createTrip: true,
			bodies:     []string{`{"status":"flying"}`},
			wantCode:   http.StatusUnprocessableEntity,
		},
		{
			name:       "missing status field",
			createTrip: true,
			bodies:     []string{`{}`},
			wantCode:   http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid JSON body",
			createTrip: true,
			bodies:     []string{`{"status":`},
			wantCode:   http.StatusBadRequest,
		},
		{
			name:     "invalid trip ID",
			id:       "abc",
			bodies:   []string{`{"status":"accepted"}`},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "trip not found",
			id:       "999",
			bodies:   []string{`{"status":"accepted"}`},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()

			id := tt.id
			if tt.createTrip {
				id = strconv.FormatInt(createTripForTest(t, h), 10)
			}

			var rec *httptest.ResponseRecorder
			for i, body := range tt.bodies {
				rec = patchStatus(t, h, id, body)
				if i < len(tt.bodies)-1 && rec.Code != http.StatusOK {
					t.Fatalf("setup step %d returned %d: %s", i, rec.Code, rec.Body.String())
				}
			}

			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d; body: %s", tt.wantCode, rec.Code, rec.Body.String())
			}

			if tt.wantStatus == "" {
				return
			}

			var body TripResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}
			if body.Status != tt.wantStatus {
				t.Errorf("expected status %q, got %q", tt.wantStatus, body.Status)
			}
		})
	}
}

// conflictRepo reproduces what FakeTripRepository cannot: the status changes
// between the service's read and its conditional update.
type conflictRepo struct{}

func (conflictRepo) Save(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	return trip, nil
}

func (conflictRepo) GetById(ctx context.Context, id int64) (domain.Trip, error) {
	return domain.Trip{ID: id, Status: domain.TripStatusRequested}, nil
}

func (conflictRepo) UpdateStatus(context.Context, int64, domain.TripStatus, domain.TripStatus) (domain.Trip, error) {
	return domain.Trip{}, domain.ErrTripStatusConflict
}

func (conflictRepo) AssignDriver(
	context.Context,
	int64,
	int64,
) (domain.Trip, error) {
	return domain.Trip{}, domain.ErrTripStatusConflict
}

func TestTripHandler_UpdateTripStatus_Conflict(t *testing.T) {
	h := NewTripHandler(service.NewTripService(conflictRepo{}))

	rec := patchStatus(t, h, "1", `{"status":"cancelled"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected %d, got %d; body: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}
