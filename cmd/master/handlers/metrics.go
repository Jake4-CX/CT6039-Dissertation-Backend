package handlers

import (
	"encoding/json"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func HandleWorkerMetrics(d amqp.Delivery) {
	var metrics structs.LoadTestWorkerMetrics
	if err := json.Unmarshal(d.Body, &metrics); err != nil {
		log.Errorf("Error unmarshalling metrics: %s", err)
		return
	}

	managers.AggregateMetrics(metrics.LoadTestID, metrics.ResponseFragments, metrics.Timestamp)
}