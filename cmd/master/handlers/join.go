package handlers

import (
	"encoding/json"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func HandleWorkerJoin(d amqp.Delivery) {
	var message struct {
		ID           string  `json:"id"`
		CPUUsage     float64 `json:"cpu_usage"`     // CPU usage percentage
		TotalMem     uint64  `json:"total_mem"`     // Total virtual memory
		AvailableMem uint64  `json:"available_mem"` // Available memory
	}
	if err := json.Unmarshal(d.Body, &message); err != nil {
		log.Errorf("Error unmarshalling worker join message: %s", err)
		return
	}

	worker := &structs.Worker{
		ID:            message.ID,
		LastHeartbeat: time.Now(),
		Capabilities: structs.Capabilities{
			CPUUsage:     message.CPUUsage,
			TotalMem:     message.TotalMem,
			AvailableMem: message.AvailableMem,
		},
		CurrentLoad: 0,
		MaxLoad:     100,
		Available:   true,
	}

	managers.AddOrUpdateWorker(message.ID, worker)
	log.Infof("Worker %s joined with capabilities: CPU Usage %f, Total Memory %d, Available Memory %d",
		message.ID, message.CPUUsage, message.TotalMem, message.AvailableMem)
}
