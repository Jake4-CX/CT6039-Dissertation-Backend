package managers

import (
	"errors"
	"sync"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type LoadTestMetricsManager struct {
	Metrics map[uint]*map[int64][]structs.ResponseFragment // LoadTestID -> Metrics
	lock    sync.RWMutex
}

var MetricsManager *LoadTestMetricsManager

func init() {
	MetricsManager = &LoadTestMetricsManager{
		Metrics: make(map[uint]*map[int64][]structs.ResponseFragment),
		lock:    sync.RWMutex{},
	}
}

func GetLoadTest(id uuid.UUID) (structs.LoadTestModel, error) {
	var loadTest structs.LoadTestModel
	result := initializers.DB.Preload("TestPlan").Preload("LoadTests").Preload("LoadTests.TestMetrics").Find(&loadTest, "UUID = ?", id)

	if result.Error != nil {
		return structs.LoadTestModel{}, errors.New("load test with id not found")
	}

	return loadTest, nil
}

func GetLoadTestsTest(id uint) (structs.LoadTestTestsModel, error) {
	var loadTest structs.LoadTestTestsModel
	result := initializers.DB.Preload("TestMetrics").First(&loadTest, id)

	if result.Error != nil {
		return structs.LoadTestTestsModel{}, errors.New("load test's test with id not found")
	}

	return loadTest, nil
}

func GetLoadTestsTestFullMetrics(testTests structs.LoadTestTestsModel) (*map[int64][]structs.ResponseFragment, error) {
	MetricsManager.lock.RLock()
	defer MetricsManager.lock.RUnlock()

	if values, exists := MetricsManager.Metrics[testTests.ID]; exists {
		return values, nil
	}

	return nil, errors.New("metrics not found")
}

func GetRunningLoadTests() ([]structs.LoadTestTestsModel, error) { // Get all running tests
	var loadTests []structs.LoadTestTestsModel
	result := initializers.DB.Preload("LoadTests").Preload("TestMetrics").Where("State = ?", structs.Running).Find(&loadTests)

	if result.Error != nil {
		return []structs.LoadTestTestsModel{}, result.Error
	}

	return loadTests, nil
}

func GetRunningLoadTestsByLoadTest(loadTest structs.LoadTestModel) []structs.LoadTestTestsModel { // Get all running tests for a particular load test
	runningTests := make([]structs.LoadTestTestsModel, 0)

	for _, loadTestTest := range loadTest.LoadTests {
		if loadTestTest.State == structs.Running {
			runningTests = append(runningTests, loadTestTest)
		}
	}

	return runningTests
}

func StartLoadTest(loadTest structs.LoadTestModel, duration int, virtualUsers int, loadTestType structs.LoadTestType) (structs.LoadTestTestsModel, error) {
	log.Infof("Starting load test with ID %s", loadTest.UUID)

	testMetrics := structs.LoadTestMetricsModel{
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		TotalResponseTime:   0,
		AverageResponseTime: 0,
	}

	if err := initializers.DB.Save(&testMetrics).Error; err != nil {
		log.Errorf("Error saving test metrics: %s", err)
		return structs.LoadTestTestsModel{}, err
	}

	newTest := structs.LoadTestTestsModel{
		LoadTestModelId: loadTest.ID,
		State:           structs.Running,
		Duration:        duration,
		VirtualUsers:    virtualUsers,
		LoadTestType:    loadTestType,
		TestMetrics:     testMetrics,
	}

	if err := initializers.DB.Save(&newTest).Error; err != nil {
		log.Errorf("Error starting load test: %s", err)
		return structs.LoadTestTestsModel{}, err
	}

	log.Infof("Load test with ID %s started", loadTest.UUID)

	// Create callback to provide to avoid import cycle
	completionCallback := func(testModel structs.LoadTestTestsModel) error {
		_, err := CompleteLoadTestByTestModel(testModel)
		return err
	}

	initializers.InitalizeTest(&newTest, loadTest.TestPlan, GetAvailableWorkers(), completionCallback)
	return newTest, nil
}

func StopLoadTest(loadTest structs.LoadTestModel) (structs.LoadTestTestsModel, error) {
	var loadTestsTest structs.LoadTestTestsModel
	updateResult := initializers.DB.Model(&loadTestsTest).Where("LoadTestModelId = ? AND State = ?", loadTest.ID, structs.Running).Update("State", structs.Cancelled)

	if updateResult.Error != nil {
		return structs.LoadTestTestsModel{}, updateResult.Error
	}

	initializers.CancelTest(loadTestsTest, GetAvailableWorkers())

	return loadTestsTest, nil
}

func StopLoadTestByTestModel(loadTestsTest structs.LoadTestTestsModel) (structs.LoadTestTestsModel, error) {
	updateResult := initializers.DB.Model(&loadTestsTest).Where("State = ?", structs.Running).Update("State", structs.Cancelled)

	if updateResult.Error != nil {
		return structs.LoadTestTestsModel{}, updateResult.Error
	}

	initializers.CancelTest(loadTestsTest, GetAvailableWorkers())

	return loadTestsTest, nil
}

func CompleteLoadTest(loadTest structs.LoadTestModel) (structs.LoadTestTestsModel, error) {
	var loadTestsTest structs.LoadTestTestsModel
	updateResult := initializers.DB.Model(&loadTestsTest).Where("LoadTestModelId = ? AND State = ?", loadTest.ID, structs.Running).Update("State", structs.Completed)

	if updateResult.Error != nil {
		return structs.LoadTestTestsModel{}, updateResult.Error
	}

	return loadTestsTest, nil
}

func CompleteLoadTestByTestModel(loadTestsTest structs.LoadTestTestsModel) (structs.LoadTestTestsModel, error) {
	updateResult := initializers.DB.Model(&loadTestsTest).Where("State = ?", structs.Running).Update("State", structs.Completed)

	if updateResult.Error != nil {
		return structs.LoadTestTestsModel{}, updateResult.Error
	}

	return loadTestsTest, nil
}

func AggregateMetrics(testId uint, responseFragments []structs.ResponseFragment, reportedAt int64) {

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

	testsTest, err := GetLoadTestsTest(testId)
	if err != nil {
		log.Errorf("Load test's test with ID %d not found for metrics aggregation (sql)", testId)
		log.Errorf("It's error: %s", err.Error())
		return
	}

	// Calculate elapsed time
	startTime := testsTest.CreatedAt.UnixNano() / int64(time.Millisecond)
	elapsedSeconds := (reportedAt - startTime) / 1000

	log.Infof("Aggregating metrics for load test with ID %d at %d seconds for timestamp %d", testId, elapsedSeconds, reportedAt)

	// Local metrics

	MetricsManager.lock.Lock()
	defer MetricsManager.lock.Unlock()

	if _, exists := MetricsManager.Metrics[testId]; !exists {
		log.Errorf("Load test with ID %d not found for metrics aggregation (local)", testId)
		MetricsManager.Metrics[testId] = &map[int64][]structs.ResponseFragment{}
	}

	// Append metrics to map
	metricsMap := *MetricsManager.Metrics[testId]
	metricsMap[elapsedSeconds] = append(metricsMap[elapsedSeconds], responseFragments...)

	// Global metrics

	totalRequests := testsTest.TestMetrics.TotalRequests + int(len(responseFragments))

	if totalRequests > 0 {
		averageResponseTime := totalResponseTime / int64(totalRequests)

		err := initializers.DB.Model(&testsTest.TestMetrics).Updates(structs.LoadTestMetricsModel{
			TotalRequests:       totalRequests,
			SuccessfulRequests:  testsTest.TestMetrics.SuccessfulRequests + totalSuccessRequests,
			FailedRequests:      testsTest.TestMetrics.FailedRequests + totalFailedRequests,
			TotalResponseTime:   totalResponseTime,
			AverageResponseTime: averageResponseTime,
		}).Error

		if err != nil {
			log.Errorf("Error updating load test metrics: %s", err)
			return
		} else {
			log.Infof("Metrics updated successfully for ID %d at %d seconds for timestamp %d", testId, elapsedSeconds, reportedAt)
		}

	} else {
		log.Error("No requests were made, skipping metrics update.")
	}

}
