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

func AggregateMetrics(id uuid.UUID, newMetrics structs.LoadTestMetricFragment) {
	LoadManager.lock.Lock()
	defer LoadManager.lock.Unlock()

	// Log the current state before updating
	if test, exists := LoadManager.LoadTests[id]; exists {
		test.Metrics.GlobalMetrics.TotalRequests += newMetrics.TotalRequests
		test.Metrics.GlobalMetrics.SuccessfulRequests += newMetrics.SuccessfulRequests
		test.Metrics.GlobalMetrics.FailedRequests += newMetrics.FailedRequests
		test.Metrics.GlobalMetrics.TotalResponseTime += newMetrics.TotalResponseTime
		test.Metrics.GlobalMetrics.AverageResponseTime = test.Metrics.GlobalMetrics.TotalResponseTime / int64(test.Metrics.GlobalMetrics.TotalRequests)
		test.LastUpdateAt = time.Now()

		test.Metrics.Metrics = append(test.Metrics.Metrics, newMetrics)

	} else {
		log.Errorf("Load test with ID %s not found for metrics aggregation", id)
	}

}
