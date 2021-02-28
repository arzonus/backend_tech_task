package application

import (
	"fmt"
	"time"

	"github.com/PxyUp/backend_tech_task/internal/external"
)

type Application struct {
	ID             string
	Status         Status
	UserID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExternalStatus external.Status
}

type Status int32

var ErrInvalidStatus = fmt.Errorf("invalid status")

func NewStatus(val int32) Status {
	return Status(val)
}

func (s Status) Validate() error {
	switch s {
	case StatusOpen, StatusInProgress, StatusClosed:
		return nil
	}
	return ErrInvalidStatus
}

func (s Status) String() string {
	switch s {
	case StatusOpen:
		return "open"
	case StatusInProgress:
		return "in_progress"
	case StatusClosed:
		return "closed"
	}
	return "unspecified"
}

func (s Status) Int32() int32 {
	return int32(s)
}

const (
	StatusUnspecified Status = iota
	StatusOpen
	StatusInProgress
	StatusClosed
)
