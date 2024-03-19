package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/initializers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

func GetLoadTests(c *gin.Context) {

	var loadTests []structs.LoadTestModel
	initializers.DB.Preload("TestPlan").Preload("LoadTests").Find(&loadTests)

	c.JSON(200, loadTests)

}

func GetLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	result, err := managers.GetLoadTest(loadTestUUID)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func CreateLoadTest(c *gin.Context) {
	var newLoadTest struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&newLoadTest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	reactFlowPlan, err := utils.LoadJSONFromFile("config/testPlans/reactFlowPlan.json")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load ReactFlowPlan"})
		return
	}

	testPlan, err := utils.LoadJSONFromFile("config/testPlans/testPlan.json")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load TestPlan"})
		return
	}

	loadTestEntry := structs.LoadTestModel{
		Name: newLoadTest.Name,
		TestPlan: structs.LoadTestPlanModel{
			ReactFlowPlan: reactFlowPlan,
			TestPlan:      testPlan,
		},
	}

	result := initializers.DB.Create(&loadTestEntry)

	if result.Error != nil {
		log.Error("Failed to create load test", result.Error.Error())
		c.JSON(500, gin.H{"error": "Failed to create load test"})
		return
	}

	// managers.NewLoadTest(loadTest.ID, loadTest.Name, loadTest.LoadTestPlan.URL, loadTest.LoadTestPlan.Duration, loadTest.LoadTestPlan.VirtualUsers)

	c.JSON(200, gin.H{"message": "Load test created", "data": loadTestEntry})
}

func DeleteLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	var loadTest structs.LoadTestModel
	result := initializers.DB.First(&loadTest, "UUID = ?", loadTestID)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": "Load test not found"})
		return
	}

	initializers.DB.Delete(&loadTest)

	c.JSON(200, gin.H{"message": "Load test deleted"})
}

func StartLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	// Parse request body
	var newLoadTestExecution struct {
		Duration     int                  `json:"duration" binding:"required"`
		VirtualUsers int                  `json:"virtualUsers" binding:"required"`
		LoadTestType structs.LoadTestType `json:"loadTestType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&newLoadTestExecution); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get load test
	loadTest, err := managers.GetLoadTest(loadTestUUID)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// Start load test

	runningTests := managers.GetRunningLoadTestsByLoadTest(loadTest)
	if len(runningTests) > 0 {
		c.JSON(400, gin.H{"error": "A Load test is already running for this test plan"})
		return
	}

	result, err := managers.StartLoadTest(loadTest, newLoadTestExecution.Duration, newLoadTestExecution.VirtualUsers, newLoadTestExecution.LoadTestType)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start load test"})
		return
	}

	c.JSON(200, gin.H{"message": "Load test started", "data": result})
}

func StopLoadTest(c *gin.Context) {
	id := c.Param("id")

	loadTestUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	// Get load test
	loadTest, err := managers.GetLoadTest(loadTestUUID)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// Stop load test
	result, err := managers.StopLoadTest(loadTest)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to stop load test", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Load test stopped", "data": result})
}

func UpdateLoadTestPlan(c *gin.Context) {
	id := c.Param("id")

	loadTestUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UUID"})
		return
	}

	var requestBody structs.UpdateTestPlanRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if load test exists

	var loadTest structs.LoadTestModel

	result := initializers.DB.Preload("TestPlan").Where("UUID = ?", loadTestUUID).First(&loadTest)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": "Load test not found"})
		return
	}

	// Update load test plan

	reactFlowPlanJSON, _ := json.Marshal(requestBody.ReactFlow)
	testPlanJSON, _ := json.Marshal(requestBody.TestPlan)

	loadTest.TestPlan.ReactFlowPlan = (string(reactFlowPlanJSON))
	loadTest.TestPlan.TestPlan = (string(testPlanJSON))

	// Save changes
	if err := initializers.DB.Save(&loadTest.TestPlan).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update test plan"})
		return
	}

	c.JSON(200, gin.H{"message": "Test plan updated"})
}
