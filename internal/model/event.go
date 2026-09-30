package model

import (
	"fmt"
	"time"
)

type Event struct {
	ID         string    `json:"id" binding:"required"`
	Type       string    `json:"type" binding:"required"`
	Status     string    `json:"status" binding:"required"`
	OccurredAt time.Time `json:"occurred_at" binding:"required"`
	ReceivedAt time.Time `json:"-"`
}

func (event *Event) Validate() error {
	if event.ID == "" {
		return fmt.Errorf("event id is required")
	}
	if event.Type == "" {
		return fmt.Errorf("event type is required")
	}
	if event.Status == "" {
		return fmt.Errorf("event status is required")
	}
	if event.OccurredAt.IsZero() {
		return fmt.Errorf("event occurred_at is required")
	}
	return nil
}
