package handlers

import (
	"encoding/json"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

func SendJoinMessage(workerID string) {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	virtualMem, _ := mem.VirtualMemory()

	message := struct {
		ID           string  `json:"id"`
		CPUUsage     float64 `json:"cpu_usage"`     // Average CPU usage percentage
		TotalMem     uint64  `json:"total_mem"`     // Total virtual memory in bytes
		AvailableMem uint64  `json:"available_mem"` // Available memory in bytes
	}{
		ID:           workerID,
		CPUUsage:     cpuPercent[0],
		TotalMem:     virtualMem.Total,
		AvailableMem: virtualMem.Available,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal join message: %s", err)
	}

	err = initializers.RabbitCh.Publish(
		"",
		"worker.join",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		})
	if err != nil {
		log.Fatalf("Failed to send join message: %s", err)
	}
}

func SendLeaveMessage(workerID string) {
	err := initializers.RabbitCh.Publish(
		"",
		"worker.leave",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(workerID),
		})
	if err != nil {
		log.Errorf("Failed to send leave message: %s", err)
	}
}

func StartHeartbeat(workerID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := initializers.RabbitCh.Publish(
				"",
				"worker.heartbeat",
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(workerID),
				})
			if err != nil {
				log.Errorf("Failed to send heartbeat: %s", err)
				return
			}
		}
	}
}
