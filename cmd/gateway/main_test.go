package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTripHTTPFlow(t *testing.T) {
	router := newRouter()

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
