package external

import (
	"fmt"
	"strings"
)

type Status int32

const (
	StatusUnspecified Status = iota
	StatusProcessed
	StatusSkipped
)

var ErrInvalidStatus = fmt.Errorf("invalid status")

func NewStatus(val int32) Status {
	return Status(val)
}

func NewStatusString(val string) (Status, error) {
	val = strings.ToLower(val)
	switch val {
	case "processed":
		return StatusProcessed, nil
	case "skipped":
		return StatusSkipped, nil
	}

	return StatusUnspecified, fmt.Errorf("%w: %s", ErrInvalidStatus, val)
}

func (s Status) Validate() error {
	switch s {
	case StatusProcessed, StatusSkipped:
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalidStatus, s.String())
}

func (s Status) String() string {
	switch s {
	case StatusProcessed:
		return "processed"
	case StatusSkipped:
		return "skipped"
	}
	return "unspecified"
}

func (s Status) Int32() int32 {
	return int32(s)
}
