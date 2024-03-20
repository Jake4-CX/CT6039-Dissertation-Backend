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
	// log.Infof("Received message: %s", string(d.Body))

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

	services.ExecuteLoadTest(assignment)

	log.Infof("Executed assigned task: %+v", assignment.LoadTestTestsModel)
}

func HandleTaskCancelled(d amqp.Delivery) {
	// log.Infof("Received message: %s", string(d.Body))

	var assignment structs.TaskAssignment
	err := json.Unmarshal(d.Body, &assignment)
	if err != nil {
		log.Errorf("Error unmarshalling task assignment: %s", err)
		return
	}

	log.Infof("Cancelling load test with ID: %d", assignment.LoadTestTestsModel.ID)

	if cancelFunc, ok := state.LoadTestCancellers.Load(assignment.LoadTestTestsModel.ID); ok {
		cancelFunc.(context.CancelFunc)()
		state.LoadTestCancellers.Delete(assignment.LoadTestTestsModel.ID)
	}
}
