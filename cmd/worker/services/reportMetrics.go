package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func ReportMetricsPeriodically(ctx context.Context, workerID string, responseChannel <-chan structs.ResponseItem, duration int, loadTestID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initialize metrics outside the loop
	var metrics structs.LoadTestWorkerMetrics

	// Listen for responses and ticks
	for {
		select {
		case <-ctx.Done():
			// Context has been canceled, stop reporting
			return
		case resp := <-responseChannel:
			// Aggregate response metrics
			metrics.TotalRequests++
			metrics.TotalResponseTime += resp.ResponseTime
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				metrics.SuccessfulRequests++
			} else {
				metrics.FailedRequests++
			}
		case <-ticker.C:
			// Time to report the metrics
			if metrics.TotalRequests > 0 {
				metrics.AverageResponseTime = metrics.TotalResponseTime / int64(metrics.TotalRequests)
			}

			// Prepare metrics for reporting
			report := structs.LoadTestWorkerMetrics{
				WorkerID:              workerID,
				LoadTestID:            loadTestID,
				LoadTestMetricFragment: metrics.LoadTestMetricFragment,
			}

			// Send the report
			metricsJSON, _ := json.Marshal(report)
			_ = initializers.RabbitCh.Publish(
				"", "worker.performance.metrics", false, false,
				amqp.Publishing{ContentType: "application/json", Body: metricsJSON},
			)

			// Log for visibility
			log.Infof("Reported metrics: %+v", report)

			// Reset metrics after reporting
			metrics = structs.LoadTestWorkerMetrics{}

			// Check if the duration has elapsed
			if time.Since(time.Now()).Milliseconds() > int64(duration) {
				return
			}
		}
	}
}

