package controllers

import (
	"net/http"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

func GetLoadTests(c *gin.Context) {

	var loadTests []structs.LoadTest = make([]structs.LoadTest, 0)

	for _, loadTest := range managers.LoadManager.LoadTests {
		loadTests = append(loadTests, *loadTest)
	}

	c.JSON(200, loadTests)

}

func GetLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	if loadTest, exists := managers.LoadManager.LoadTests[loadTestID]; exists {
		c.JSON(200, loadTest)
		return
	} else {
		c.JSON(404, gin.H{"error": "Load test not found"})
		return
	}
}

func CreateLoadTest(c *gin.Context) {
	var newLoadTest struct {
		Name         string `json:"name" binding:"required"`
		URL          string `json:"url" binding:"required"`
		Duration     int    `json:"duration" binding:"required"`
		VirtualUsers int    `json:"virtualUsers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&newLoadTest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var loadTest = structs.LoadTest{
		ID:           uuid.New(),
		Name:         newLoadTest.Name,
		State:        structs.Pending,
		CreatedAt:    time.Now(),
		LastUpdateAt: time.Now(),
		Metrics:      structs.LoadTestMetrics{},
		LoadTestPlan: structs.LoadTestPlan{
			URL:          newLoadTest.URL,
			Duration:     newLoadTest.Duration,
			VirtualUsers: newLoadTest.VirtualUsers,
		},
	}

	managers.NewLoadTest(loadTest.ID, loadTest.Name, loadTest.LoadTestPlan.URL, loadTest.LoadTestPlan.Duration, loadTest.LoadTestPlan.VirtualUsers)

	c.JSON(201, loadTest)
}

func StartLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	loadTest, exists := managers.LoadManager.LoadTests[loadTestID]
	if !exists {
		c.JSON(404, gin.H{"error": "Load test not found"})
		return
	}

	if loadTest.State == structs.Pending || loadTest.State == structs.Completed || loadTest.State == structs.Cancelled {
		managers.UpdateLoadTestState(loadTestID, structs.Running)
		c.JSON(200, gin.H{"message": "Load test started"})

		return
	} else {
		c.JSON(400, gin.H{"error": "Load test is not in a pending state"})

		return
	}
}

func StopLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	if loadTest, exists := managers.LoadManager.LoadTests[loadTestID]; exists {

		if loadTest.State == structs.Running {
			// ToDo
			managers.UpdateLoadTestState(loadTestID, structs.Cancelled)
			c.JSON(200, gin.H{"message": "Load test stopped"})
			return
		} else {
			c.JSON(400, gin.H{"error": "Load test is not in a running state"})
			return
		}
	} else {
		c.JSON(404, gin.H{"error": "Load test not found"})
		return
	}
}

func UpdateLoadTestPlan(c *gin.Context) {
	id := c.Param("id")

	loadTestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	var requestBody structs.UpdateTestPlanRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Info("Test plan: ", requestBody.TestPlan)
	log.Info("Load test ID: ", loadTestID)
	log.Info("ReactFlow Edges: ", requestBody.ReactFlow.Edges)

	c.JSON(200, gin.H{"message": "Test plan updated"})
}
