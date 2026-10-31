package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ride-sharing/internal/trip/handler"
	"ride-sharing/internal/trip/service"
	"strings"
	"testing"
)

func TestTripHTTPFlow(t *testing.T) {
	fakeRepo := service.NewFakeTripRepository()
	svc := service.NewTripService(fakeRepo)

	th := handler.NewTripHandler(svc)

	router := newRouter(th)

	// Create a trip.
	createReq := httptest.NewRequest(
		http.MethodPost,
		"/trips",
		strings.NewReader(
			`{"pickup":"Berlin","destination":"Hamburg"}`,
		),
	)
	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf(
			"expected create status %d, got %d: %s",
			http.StatusCreated,
			createRec.Code,
			createRec.Body.String(),
		)
	}

	var created struct {
		ID          int    `json:"id"`
		Pickup      string `json:"pickup"`
		Destination string `json:"destination"`
	}

	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response JSON: %v", err)
	}

	// Retrieve the created trip.
	getReq := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/trips/%d", created.ID),
		nil,
	)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf(
			"expected get status %d, got %d: %s",
			http.StatusOK,
			getRec.Code,
			getRec.Body.String(),
		)
	}

	var retrieved struct {
		ID          int    `json:"id"`
		Pickup      string `json:"pickup"`
		Destination string `json:"destination"`
	}

	if err := json.Unmarshal(getRec.Body.Bytes(), &retrieved); err != nil {
		t.Fatalf("invalid get response JSON: %v", err)
	}

	if retrieved.ID != created.ID ||
		retrieved.Pickup != "Berlin" ||
		retrieved.Destination != "Hamburg" {
		t.Fatalf("unexpected retrieved trip: %+v", retrieved)
	}
}

func TestTripStatusHTTPFlow(t *testing.T) {
	fakeRepo := service.NewFakeTripRepository()
	svc := service.NewTripService(fakeRepo)
	router := newRouter(handler.NewTripHandler(svc))

	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, httptest.NewRequest(
		http.MethodPost,
		"/trips",
		strings.NewReader(`{"pickup":"Berlin","destination":"Hamburg"}`),
	))
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
	}

	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response JSON: %v", err)
	}

	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/trips/%d/status", created.ID),
		strings.NewReader(`{"status":"cancelled"}`),
	))
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch: expected %d, got %d: %s", http.StatusOK, patchRec.Code, patchRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/trips/%d", created.ID),
		nil,
	))
	if getRec.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, getRec.Code, getRec.Body.String())
	}

	var retrieved struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &retrieved); err != nil {
		t.Fatalf("invalid get response JSON: %v", err)
	}

	if retrieved.Status != "cancelled" {
		t.Fatalf("expected status %q, got %q", "cancelled", retrieved.Status)
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	fakeRepo := service.NewFakeTripRepository()
	svc := service.NewTripService(fakeRepo)
	router := newRouter(handler.NewTripHandler(svc))

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "GET on create endpoint", method: http.MethodGet, path: "/trips"},
		{name: "PUT on status endpoint", method: http.MethodPut, path: "/trips/1/status"},
		{name: "DELETE on get endpoint", method: http.MethodDelete, path: "/trips/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`)))

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
			}
		})
	}
}
