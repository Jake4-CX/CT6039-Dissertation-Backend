package initializers

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
)

type CompletionCallback func(testModel structs.LoadTestTestsModel) error

func InitalizeTest(loadTest *structs.LoadTestTestsModel, testPlan structs.LoadTestPlanModel, availableWorkers []*structs.Worker, onComplete CompletionCallback) {

	log.Infof("Initializing load test with ID %d", loadTest.ID)

	// split strategy
	virtualUsersPerWorker := loadTest.VirtualUsers / len(availableWorkers)
	remainingUsers := loadTest.VirtualUsers % len(availableWorkers)

	for _, worker := range availableWorkers {
		log.Infof("Assigning task to worker %s", worker.ID)
		virtualUsersForThisWorker := virtualUsersPerWorker
		if remainingUsers > 0 {
			virtualUsersForThisWorker++
			remainingUsers--
		}

		sendStartTaskToWorker(*loadTest, testPlan, worker.ID)
	}

	go func() {
		duration := time.Duration(loadTest.Duration) * time.Millisecond
		time.Sleep(duration)

		onComplete(*loadTest)

		log.Infof("Test duration reached. Stopping load test ID %d", loadTest.ID)
	}()
}

func CancelTest(loadTestsTest structs.LoadTestTestsModel, availableWorkers []*structs.Worker) {
	// ToDo: Method to check if the loadTest Task has been assigned to the workers in question

	for _, worker := range availableWorkers {
		sendStopTaskToWorker(loadTestsTest, worker.ID)
	}
}

func sendStartTaskToWorker(loadTest structs.LoadTestTestsModel, testPlan structs.LoadTestPlanModel, workerID string) {

	assignment := structs.TaskAssignment{
		LoadTestTestsModel: loadTest,
		LoadTestPlanModel:  testPlan,
		AssignedWorkerID:   workerID,
	}

	taskBytes, err := json.Marshal(assignment)
	if err != nil {
		log.Errorf("Failed to marshal task assignment: %s", err)
		return
	}

	err = initializers.RabbitCh.Publish(
		"",            // exchange
		"task.create", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskBytes,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish task assignment: %s", err)
	}

	log.Infof("Task created!")
}

func sendStopTaskToWorker(loadTest structs.LoadTestTestsModel, workerID string) {

	cancelAssignment := structs.TaskAssignment{
		LoadTestTestsModel: loadTest,
		AssignedWorkerID:   workerID,
	}

	msgBytes, err := json.Marshal(cancelAssignment)
	if err != nil {
		log.Errorf("Failed to marshal cancellation message: %s", err)
		return
	}

	err = initializers.RabbitCh.Publish(
		"",            // exchange
		"task.cancel", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		log.Errorf("Failed to publish load test cancellation message: %s", err)
	}

}

// func getAvailableWorkers() []*structs.Worker {
// 	var availableWorkers []*structs.Worker
// 	for _, worker := range workers {
// 		if worker.Available && (worker.CurrentLoad < worker.MaxCapacity) {
// 			availableWorkers = append(availableWorkers, worker)
// 		}
// 	}
// 	return availableWorkers
// }
