package handlers

import (
	"context"
	"encoding/json"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/services"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/state"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func HandleTaskCreated(d amqp.Delivery) {
	log.Infof("Received message: %s", string(d.Body))

	var assignment structs.TaskAssignment
	err := json.Unmarshal(d.Body, &assignment)
	if err != nil {
		log.Errorf("Error unmarshalling task assignment: %s", err)
		return
	}

	if assignment.AssignedWorkerID != WorkerID {
		log.Debugf("Task not assigned to this worker (%s). Ignoring.", WorkerID)
		return
	}

	log.Infof("Executing assigned task: %+v", assignment.Task)

	metrics := services.ExecuteLoadTest(assignment.Task, WorkerID, assignment.LoadTestID)

	log.Infof("Load test metrics: %+v", metrics)

	log.Infof("Executed assigned task: %+v", assignment.Task)
}

func HandleTaskCancelled(d amqp.Delivery) {
	log.Infof("Received message: %s", string(d.Body))

	var loadTest struct {
		LoadTestID       string `json:"loadTestId"`
		AssignedWorkerID string `json:"assignedWorkerId"`
	}
	err := json.Unmarshal(d.Body, &loadTest)
	if err != nil {
		log.Errorf("Error unmarshalling load test ID: %s", err)
		return
	}

	log.Infof("Cancelling load test with ID: %s", loadTest.LoadTestID)

	if cancelFunc, ok := state.LoadTestCancellers.Load(loadTest.LoadTestID); ok {
		cancelFunc.(context.CancelFunc)()
		state.LoadTestCancellers.Delete(loadTest.LoadTestID)
	}
}
