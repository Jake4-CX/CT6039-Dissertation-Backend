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

func ReportMetricsPeriodically(ctx context.Context, workerID string, responseChannel <-chan structs.ResponseItem, duration int, loadTestID uint) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	var responseFragments []structs.ResponseFragment = make([]structs.ResponseFragment, 0)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-responseChannel:
				responseFragments = append(responseFragments, structs.ResponseFragment{StatusCode: resp.StatusCode, ResponseTime: resp.ResponseTime})
			}
		}

	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Time to report metrics

			fragments := &responseFragments

			metrics := structs.LoadTestWorkerMetrics{
				WorkerID:          workerID,
				LoadTestID:        loadTestID,
				Timestamp:         time.Now().UnixNano() / int64(time.Millisecond),
				ResponseFragments: *fragments,
			}

			responseFragments = make([]structs.ResponseFragment, 0)

			metricsJSON, _ := json.Marshal(metrics)
			_ = initializers.RabbitCh.Publish(
				"", "worker.performance.metrics", false, false,
				amqp.Publishing{ContentType: "application/json", Body: metricsJSON},
			)

			log.Debugf("Reported metrics: %+v", metrics)

			if time.Since(time.Now()).Milliseconds() > int64(duration) {
				log.Info("Duration elapsed")
				return
			}
		}
	}

}
