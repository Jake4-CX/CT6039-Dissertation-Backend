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
	ID           uuid.UUID       `json:"id"` // A unique identifier for the load test
	Name         string          `json:"name"`
	State        LoadTestState   `json:"state"`
	CreatedAt    time.Time       `json:"createdAt"`
	LastUpdateAt time.Time       `json:"lastUpdateAt"`
	Metrics      LoadTestMetrics `json:"metrics"`
	LoadTestPlan LoadTestPlan    `json:"loadTestPlan"`
}

type LoadTestPlan struct {
	URL          string `json:"url"`
	Duration     int    `json:"duration"`
	VirtualUsers int    `json:"virtualUsers"`
}

type LoadTestMetrics struct {
	GlobalMetrics LoadTestMetricFragment   `json:"globalMetrics"`
	Metrics       []LoadTestMetricFragment `json:"metrics"`
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
