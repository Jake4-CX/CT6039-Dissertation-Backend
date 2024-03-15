package handlers

import (
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func HandleWorkerLeave(d amqp.Delivery) {
	workerID := string(d.Body)
	
	managers.RemoveWorker(workerID)
	log.Infof("Worker left: %s", workerID)
}
