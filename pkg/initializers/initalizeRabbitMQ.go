package initializers

import (
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
)

var RabbitConn *amqp.Connection
var RabbitCh *amqp.Channel

func InitializeRabbitMQ() {
	var err error
	connStr := `amqp://` + os.Getenv("RABBIT_USER") + `:` + os.Getenv("RABBIT_PASS") + `@` + os.Getenv("RABBIT_HOST") + `:` + os.Getenv("RABBIT_PORT") + "/"
	RabbitConn, err = amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	RabbitCh, err = RabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	queues := []string{
		"task.create", "task.status.update", "task.delete", "task.get",
		"worker.join", "worker.leave",
		"worker.status.update", "worker.performance.metrics",
		"worker.heartbeat"}
	for _, q := range queues {
		_, err = RabbitCh.QueueDeclare(
			q,
			false, // Durable
			false, // Delete when unused
			false, // Exclusive
			false, // No-wait
			nil,   // Arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare %s queue: %s", q, err)
		}
	}

	log.Info("Successfully connected to RabbitMQ and declared queues")
}

func CleanupRabbitMQ() {
	if RabbitCh != nil {
		RabbitCh.Close()
	}
	if RabbitConn != nil {
		RabbitConn.Close()
	}
}
