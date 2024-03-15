package structs

import (
	"time"

	"github.com/google/uuid"
)

type ResponseItem struct {
	StatusCode   int
	ResponseTime int64
	Body         string
}

type LoadTestState string

const (
	Pending   LoadTestState = "PENDING"
	Running   LoadTestState = "RUNNING"
	Cancelled LoadTestState = "CANCELLED"
	Completed LoadTestState = "COMPLETED"
)

type LoadTest struct {
	ID           uuid.UUID // A unique identifier for the load test
	Name         string
	State        LoadTestState
	CreatedAt    time.Time
	LastUpdateAt time.Time
	Metrics      LoadTestMetrics
	LoadTestPlan LoadTestPlan
}

type LoadTestPlan struct {
	URL          string
	Duration     int
	VirtualUsers int
}

type LoadTestMetrics struct {
	GlobalMetrics LoadTestMetricFragment
	Metrics       []LoadTestMetricFragment
}

type LoadTestWorkerMetrics struct {
	WorkerID   string `json:"workerId"`
	LoadTestID string `json:"loadTestId"`
	LoadTestMetricFragment
}

type LoadTestMetricFragment struct {
	TotalRequests       int   `json:"totalRequests"`
	SuccessfulRequests  int   `json:"successfulRequests"`
	FailedRequests      int   `json:"failedRequests"`
	TotalResponseTime   int64 `json:"totalResponseTime"`
	AverageResponseTime int64 `json:"averageResponseTime"`
}
