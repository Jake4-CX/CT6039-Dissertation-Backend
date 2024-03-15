package initializers

import (
	"encoding/json"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
)

func InitalizeTest(loadTest *structs.LoadTest, availableWorkers []*structs.Worker) {

	// split strategy
	virtualUsersPerWorker := loadTest.LoadTestPlan.VirtualUsers / len(availableWorkers)
	remainingUsers := loadTest.LoadTestPlan.VirtualUsers % len(availableWorkers)

	for _, worker := range availableWorkers {
		log.Infof("Assigning task to worker %s", worker.ID)
		virtualUsersForThisWorker := virtualUsersPerWorker
		if remainingUsers > 0 {
			virtualUsersForThisWorker++
			remainingUsers--
		}

		task := structs.Task{
			URL:          loadTest.LoadTestPlan.URL,
			Duration:     loadTest.LoadTestPlan.Duration,
			VirtualUsers: virtualUsersForThisWorker,
		}

		sendStartTaskToWorker(task, worker.ID, loadTest.ID)
	}
}

func CancelTest(loadTestID uuid.UUID, availableWorkers []*structs.Worker) {
	// ToDo: Method to check if the loadTest Task has been assigned to the workers in question

	for _, worker := range availableWorkers {
		sendStopTaskToWorker(loadTestID.String(), worker.ID)
	}
}

func sendStartTaskToWorker(task structs.Task, workerID string, loadTestID uuid.UUID) {
	assignment := structs.TaskAssignment{
		Task:             task,
		AssignedWorkerID: workerID,
		LoadTestID:       loadTestID.String(),
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

func sendStopTaskToWorker(loadTestID string, workerID string) {

	cancellationMsg := struct {
		LoadTestID       string `json:"loadTestId"`
		AssignedWorkerID string `json:"assignedWorkerId"`
	}{
		LoadTestID:       loadTestID,
		AssignedWorkerID: workerID,
	}

	msgBytes, err := json.Marshal(cancellationMsg)
	if err != nil {
		log.Errorf("Failed to marshal cancellation message: %s", err)
		return
	}

	err = initializers.RabbitCh.Publish(
		"",            // exchange
		"task.delete", // routing key
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
