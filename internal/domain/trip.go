package domain

import (
	"errors"
)

type TripStatus string

var ErrUnknown = errors.New("something went wrong")

var ErrInvalidTripInput = errors.New("pickup or destination required")
var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTripID = errors.New("invalid trip ID")

var ErrInvalidTripStatus = errors.New("invalid trip status")
var ErrInvalidTripTransition = errors.New("invalid trip status transition")
var ErrTripStatusConflict = errors.New("trip status changed concurrently")

const (
	TripStatusRequested  TripStatus = "requested"
	TripStatusAccepted   TripStatus = "accepted"
	TripStatusInProgress TripStatus = "in_progress"
	TripStatusCompleted  TripStatus = "completed"
	TripStatusCancelled  TripStatus = "cancelled"
)

type Trip struct {
	ID          int64
	Pickup      string
	Destination string
	Status      TripStatus
}
