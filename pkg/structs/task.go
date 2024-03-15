package structs

import (
	"github.com/google/uuid"
)

type LoadTask struct {
	ID           uuid.UUID `json:"id"`
	RequestURL   string    `json:"requestURL"`
	VirtualUsers int       `json:"virtualUsers"`
	Duration     int       `json:"duration"` // in milliseconds
}

type TaskAssignment struct {
	Task
	AssignedWorkerID string    `json:"assignedWorkerId"`
	LoadTestID       string `json:"loadTestId"`
}

type Task struct {
	URL          string `json:"url"`
	Duration     int    `json:"duration"`     // Duration in seconds
	VirtualUsers int    `json:"virtualUsers"` // Amount of concurrent users
}
