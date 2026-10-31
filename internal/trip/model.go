package trip

import (
	"errors"
)

type Status string

var ErrUnknown = errors.New("something went wrong")

var ErrInvalidTripInput = errors.New("pickup or destination required")
var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTripID = errors.New("invalid trip ID")

var ErrInvalidTripStatus = errors.New("invalid trip status")
var ErrInvalidTripTransition = errors.New("invalid trip status transition")
var ErrTripStatusConflict = errors.New("trip status changed concurrently")

var ErrDriverNotFound = errors.New("driver not found")
var ErrDriverUnavailable = errors.New("driver unavailable")

const (
	StatusRequested  Status = "requested"
	StatusAccepted   Status = "accepted"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Trip struct {
	ID          int64
	Pickup      string
	Destination string
	Status      Status
	DriverID    *int64
}
