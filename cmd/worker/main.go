package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/handlers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.InitializeRabbitMQ()
	defer initializers.CleanupRabbitMQ()

	handlers.WorkerID = uuid.New().String()

	handlers.SendJoinMessage(handlers.WorkerID)
	go handlers.StartHeartbeat(handlers.WorkerID, 20*time.Second)

	startConsumer("task.create", handlers.HandleTaskCreated)
	startConsumer("task.cancel", handlers.HandleTaskCancelled)

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	handlers.SendLeaveMessage(handlers.WorkerID)
	log.Info("Worker shutting down")
}

func startConsumer(queueName string, handlerFunc func(amqp.Delivery)) {
	msgs, err := initializers.RabbitCh.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer for %s: %s", queueName, err)
	}

	go func() {
		log.Infof("Started consuming messages from %s", queueName)
		for d := range msgs {
			handlerFunc(d)
		}
	}()
}

