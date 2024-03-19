package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/state"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
)

// LoadTestExecutorService is a service that executes a load test

func ExecuteLoadTest(assignment structs.TaskAssignment) structs.LoadTestMetricSummary {
	log.Infof("Executing load test with config: %+v", assignment)

	if assignment.LoadTestTestsModel.Duration < 1000 {
		log.Warnf("Duration too short, adjusting to minimum of 1000 milliseconds")
		assignment.LoadTestTestsModel.Duration = 1000
	}
	if assignment.LoadTestTestsModel.VirtualUsers <= 0 {
		log.Errorf("VirtualUsers must be greater than 0")
		return structs.LoadTestMetricSummary{}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var wg sync.WaitGroup
	responseChannel := make(chan structs.ResponseItem, 100)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(assignment.LoadTestTestsModel.Duration)*time.Millisecond)
	defer cancel()

	state.LoadTestCancellers.Store(assignment.LoadTestTestsModel.ID, cancel)

	// Start reporting metrics periodically
	go ReportMetricsPeriodically(ctx, assignment.AssignedWorkerID, responseChannel, assignment.LoadTestTestsModel.Duration, assignment.LoadTestTestsModel.ID)

	testStartTime := time.Now()

	for i := 0; i < assignment.LoadTestTestsModel.VirtualUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Test duration is over.
					return
				default:
					// Make request
					makeAsyncRequest(ctx, client, "https://beta.kickable.net", responseChannel)
					time.Sleep(1 * time.Second) // Wait for 1 second before making the next request.
				}
			}
		}()
	}

	wg.Wait()
	close(responseChannel)

	metrics := collectMetrics(responseChannel)
	testDuration := time.Since(testStartTime)

	log.Infof("Load test completed in %s. Metrics: %+v", testDuration, metrics)
	return metrics
}

func collectMetrics(responseChannel <-chan structs.ResponseItem) structs.LoadTestMetricSummary {
	var metrics structs.LoadTestMetricSummary
	for item := range responseChannel {
		metrics.TotalRequests++
		metrics.TotalResponseTime += item.ResponseTime
		if item.StatusCode >= 200 && item.StatusCode < 300 {
			metrics.SuccessfulRequests++
		} else {
			metrics.FailedRequests++
		}
	}

	if metrics.TotalRequests > 0 {
		metrics.AverageResponseTime = metrics.TotalResponseTime / int64(metrics.TotalRequests)
	}
	return metrics
}

func makeAsyncRequest(ctx context.Context, client *http.Client, url string, responseChannel chan<- structs.ResponseItem) {

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Errorf("Failed to create request: %s", err)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: 0}
		return
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		log.Errorf("Request to %s failed: %s", url, err)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: elapsed.Milliseconds()}
		return
	}
	defer resp.Body.Close()

	log.Infof("Request to %s completed in %s with status code %d", url, elapsed, resp.StatusCode)

	responseChannel <- structs.ResponseItem{
		StatusCode:   resp.StatusCode,
		ResponseTime: elapsed.Milliseconds(),
	}
}
