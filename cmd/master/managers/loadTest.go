package managers

import (
	"sync"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type LoadTestLoadManager struct {
	LoadTests map[uuid.UUID]*structs.LoadTest
	lock      sync.RWMutex
}

var LoadManager *LoadTestLoadManager

func init() {
	LoadManager = &LoadTestLoadManager{
		LoadTests: make(map[uuid.UUID]*structs.LoadTest),
		lock:      sync.RWMutex{},
	}
}

func NewLoadTest(id uuid.UUID, loadTestName string, url string, duration, virtualUsers int) {

	LoadManager.lock.Lock()
	defer LoadManager.lock.Unlock()

	LoadManager.LoadTests[id] = &structs.LoadTest{
		ID:           id,
		Name:         loadTestName,
		State:        structs.Pending,
		CreatedAt:    time.Now(),
		LastUpdateAt: time.Now(),
		Metrics:      structs.LoadTestMetrics{},
		LoadTestPlan: structs.LoadTestPlan{
			URL:          url,
			Duration:     duration,
			VirtualUsers: virtualUsers,
		},
	}

	log.Infof("New load test created with ID: %s", id.String())
}

// ToDo: trigger this
func UpdateLoadTestState(id uuid.UUID, state structs.LoadTestState) {
	LoadManager.lock.Lock()
	defer LoadManager.lock.Unlock()

	if test, exists := LoadManager.LoadTests[id]; exists {

		if test.State == state {
			log.Warnf("Load test with ID %s is already in state %s", id, state)
			return
		}

		switch state {
		case structs.Running:
			test.State = state
			test.LastUpdateAt = time.Now()
			test.LoadTestPlan.StartedAt = time.Now()

			initializers.InitalizeTest(test, GetAvailableWorkers())

			log.Infof("Load test with ID %s is now running", id)

		case structs.Cancelled:
			test.State = state
			test.LastUpdateAt = time.Now()

			initializers.CancelTest(id, GetAvailableWorkers())

			log.Infof("Load test with ID %s has been cancelled", id)

		case structs.Completed:
			test.State = state
			test.LastUpdateAt = time.Now()
			log.Infof("Load test with ID %s has completed", id)
		case structs.Pending:
			test.State = state
			test.LastUpdateAt = time.Now()
			log.Infof("Load test with ID %s is now pending", id)
		default:
			log.Errorf("Invalid state %s for load test with ID %s", state, id)
		}
	}
}

func AggregateMetrics(id uuid.UUID, responseFragments []structs.ResponseFragment, reportedAt int64) {

	var totalSuccessRequests int
	var totalFailedRequests int

	var totalResponseTime int64

	// Preprocessing:
	for _, fragment := range responseFragments {
		if fragment.StatusCode >= 200 && fragment.StatusCode < 300 {
			totalSuccessRequests++
		} else {
			totalFailedRequests++
		}
		totalResponseTime += fragment.ResponseTime
	}

	LoadManager.lock.Lock()
	defer LoadManager.lock.Unlock()

	if test, exists := LoadManager.LoadTests[id]; exists {

		// Calculate elapsed time
		startTime := test.LoadTestPlan.StartedAt.UnixNano() / int64(time.Millisecond)
		elapsedSeconds := (reportedAt - startTime) / 1000

		log.Infof("Aggregating metrics for load test with ID %s at %d seconds for timestamp %d", id, elapsedSeconds, reportedAt)

		if test.Metrics.Metrics == nil {
			test.Metrics.Metrics = make(map[int64][]structs.ResponseFragment)
		}

		// append metrics (as there can be multiple workers reporting at the same time)
		test.Metrics.Metrics[elapsedSeconds] = append(test.Metrics.Metrics[elapsedSeconds], responseFragments...)

		totalRequests := test.Metrics.GlobalMetrics.TotalRequests + int(len(responseFragments))
		totalResponseTime := test.Metrics.GlobalMetrics.TotalResponseTime + totalResponseTime

		newMetrics := structs.LoadTestMetricSummary{
			TotalRequests:       totalRequests,
			SuccessfulRequests:  test.Metrics.GlobalMetrics.SuccessfulRequests + totalSuccessRequests,
			FailedRequests:      test.Metrics.GlobalMetrics.FailedRequests + totalFailedRequests,
			TotalResponseTime:   totalResponseTime,
			AverageResponseTime: totalResponseTime / int64(totalRequests),
		}

		test.Metrics.GlobalMetrics = newMetrics
		test.LastUpdateAt = time.Now()

	} else {
		log.Errorf("Load test with ID %s not found for metrics aggregation", id)
	}

}
