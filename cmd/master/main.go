package main

import (
	"os"
	"time"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/http/controllers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/handlers"
	pkgInitalizer "github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var workers map[string]*structs.Worker

func main() {

	pkgInitalizer.LoadEnvVariables()
	initializers.InitializeDB()
	pkgInitalizer.InitializeRabbitMQ()
	defer pkgInitalizer.CleanupRabbitMQ()

	workers = make(map[string]*structs.Worker)

	startConsumer("worker.join", handlers.HandleWorkerJoin)
	startConsumer("worker.leave", handlers.HandleWorkerLeave)
	startConsumer("worker.heartbeat", handlers.HandleHeartbeat)
	startConsumer("worker.performance.metrics", handlers.HandleWorkerMetrics)

	log.Info("Master node connected to RabbitMQ and started consuming messages")

	go monitorWorkers()

	router := gin.Default()

	router.Use(cors.New(
		cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge: 				 12 * time.Hour,
		},
	))

	// Workers
	router.GET("/load-workers", controllers.GetWorkers)


	// Load tests
	// Get load tests
	router.GET("/load-tests", controllers.GetLoadTests)
	router.GET("/load-tests/:id", controllers.GetLoadTest)

	// Start load test
	router.POST("/load-tests/:id/start", controllers.StartLoadTest)
	router.GET("/load-tests/:id/stop", controllers.StopLoadTest)

	// Create load test
	router.POST("/load-tests", controllers.CreateLoadTest)
	router.PUT("/load-tests/:id/plan", controllers.UpdateLoadTestPlan)

	// Delete load test
	router.DELETE("/load-tests/:id", controllers.DeleteLoadTest)

	log.Fatal(router.Run("0.0.0.0:" + os.Getenv("REST_PORT")))

	log.Info("Master node started. Waiting for workers...")
}

func monitorWorkers() {
	for {
		time.Sleep(20 * time.Second)
		now := time.Now()
		for id, worker := range workers {
			if now.Sub(worker.LastHeartbeat) > 40*time.Second {
				log.Warnf("Worker %s is considered disconnected due to missed heartbeat.", id)
				delete(workers, id)
			}
		}
	}
}

func startConsumer(queueName string, handlerFunc func(amqp.Delivery)) {
	msgs, err := pkgInitalizer.RabbitCh.Consume(
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
		for d := range msgs {
			handlerFunc(d)
		}
	}()
}
