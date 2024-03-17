package handlers

import (
	"encoding/json"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func HandleWorkerMetrics(d amqp.Delivery) {
	var metrics structs.LoadTestWorkerMetrics
	if err := json.Unmarshal(d.Body, &metrics); err != nil {
		log.Errorf("Error unmarshalling metrics: %s", err)
		return
	}

	loadTestID, err := uuid.Parse(metrics.LoadTestID)
	if err != nil {
		log.Errorf("Failed to parse load test ID: %s", err)
		return
	}

	managers.AggregateMetrics(loadTestID, metrics.ResponseFragments, metrics.Timestamp)
}