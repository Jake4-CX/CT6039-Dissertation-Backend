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
	LoadTestTestsModel LoadTestTestsModel `json:"loadTest"` // Contains VirtualUsers, Duration, LoadTestType
	LoadTestPlanModel  LoadTestPlanModel  `json:"testPlan"` // Contains the test plan
	AssignedWorkerID   string             `json:"assignedWorkerId"`
}

type Task struct {
	URL          string `json:"url"`
	Duration     int    `json:"duration"`     // Duration in seconds
	VirtualUsers int    `json:"virtualUsers"` // Amount of concurrent users
}
