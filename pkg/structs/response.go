package structs

type ResponseItem struct {
	StatusCode   int
	ResponseTime int64
	Body         string
}

type LoadTestWorkerMetrics struct {
	WorkerID          string             `json:"workerId"`
	LoadTestID        uint               `json:"loadTestId"`
	Timestamp         int64              `json:"timestamp"`
	ResponseFragments []ResponseFragment `json:"loadTestMetricFragments"`
}

type ResponseFragment struct {
	StatusCode   int   `json:"statusCode"`
	ResponseTime int64 `json:"responseTime"`
}

type TestHistoryFragment struct {
	Requests            int64 `json:"requests"`
	AverageResponseTime int64 `json:"averageResponseTime"`
}
