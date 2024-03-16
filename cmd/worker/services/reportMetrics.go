package services

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func ReportMetricsPeriodically(ctx context.Context, workerID string, responseChannel <-chan structs.ResponseItem, duration int, loadTestID string) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	var totalSuccessRequests int64
	var totalFailedRequests int64
	var totalResponseTime int64

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-responseChannel:

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&totalSuccessRequests, 1)
				} else {
					atomic.AddInt64(&totalFailedRequests, 1)
				}

				atomic.AddInt64(&totalResponseTime, resp.ResponseTime)
			}
		}
	
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Time to report metrics
			successSnapshot := atomic.SwapInt64(&totalSuccessRequests, 0)
			failedSnapshot := atomic.SwapInt64(&totalFailedRequests, 0)
			responseTimeSnapshot := atomic.SwapInt64(&totalResponseTime, 0)

			metrics := structs.LoadTestWorkerMetrics{
				WorkerID:   workerID,
				LoadTestID: loadTestID,
				LoadTestMetricFragment: structs.LoadTestMetricFragment{
					TotalRequests:       int(successSnapshot + failedSnapshot),
					SuccessfulRequests:  int(successSnapshot), 
					FailedRequests:      0,
					TotalResponseTime:   responseTimeSnapshot,
					AverageResponseTime: 0,
				},
			}
			if metrics.TotalRequests > 0 {
				metrics.AverageResponseTime = responseTimeSnapshot / int64(metrics.TotalRequests)
			}

			metricsJSON, _ := json.Marshal(metrics)
			_ = initializers.RabbitCh.Publish(
				"", "worker.performance.metrics", false, false,
				amqp.Publishing{ContentType: "application/json", Body: metricsJSON},
			)


			log.Infof("Reported metrics: %+v", metrics)

			if time.Since(time.Now()).Milliseconds() > int64(duration) {
				log.Info("Duration elapsed")
				return
			}
		}
	}

}

