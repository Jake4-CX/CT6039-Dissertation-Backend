package handlers

import (
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

func HandleHeartbeat(d amqp.Delivery) {
	workerID := string(d.Body)

	if worker, exists := managers.GetWorker(workerID); exists {
		worker.LastHeartbeat = time.Now()

		// Use workers manager to update the worker
		managers.AddOrUpdateWorker(workerID, worker)
		log.Infof("Heartbeat received from: %s", workerID)
	}
}
