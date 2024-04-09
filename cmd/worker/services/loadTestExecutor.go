package services

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/state"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
)

func ExecuteLoadTest(assignment structs.TaskAssignment) {
	log.Infof("Executing load test with config: %+v", assignment)

	if assignment.LoadTestTestsModel.Duration < 1000 {
		log.Warnf("Duration too short, adjusting to minimum of 1000 milliseconds")
		assignment.LoadTestTestsModel.Duration = 1000
	}
	if assignment.LoadTestTestsModel.VirtualUsers <= 0 {
		log.Errorf("VirtualUsers must be greater than 0")
		return
	}

	var wg sync.WaitGroup
	responseChannel := make(chan structs.ResponseItem, 100)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(assignment.LoadTestTestsModel.Duration)*time.Millisecond)
	defer cancel()

	state.LoadTestCancellers.Store(assignment.LoadTestTestsModel.ID, cancel)

	// Start reporting metrics periodically
	go ReportMetricsPeriodically(ctx, assignment.AssignedWorkerID, responseChannel, assignment.LoadTestTestsModel.Duration, assignment.LoadTestTestsModel.ID)

	testStartTime := time.Now()

	switch assignment.LoadTestTestsModel.LoadTestType {
	case structs.Load, structs.Soak:
		log.Info("Executing constant load test")
		executeConstantLoad(ctx, assignment, &wg, responseChannel)

	case structs.Stress:
		log.Info("Executing stress test")
		executeGradualIncreaseLoad(ctx, assignment, &wg, responseChannel)

	case structs.Spike:
		log.Info("Executing spike test")
		executeSpikeLoad(ctx, assignment, &wg, responseChannel)

	default:
		log.Errorf("Unknown load test type: %s", assignment.LoadTestTestsModel.LoadTestType)
	}

	wg.Wait()
	close(responseChannel)

	testDuration := time.Since(testStartTime)

	log.Infof("Load test completed in %s.", testDuration)
}

// Put Load Test Type Executor functions here
func executeConstantLoad(ctx context.Context, assignment structs.TaskAssignment, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {
	for i := 0; i < assignment.LoadTestTestsModel.VirtualUsers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Test duration is over.
					return
				default:
					// Make request
					simulateVirtualUser(ctx, assignment, id, responseChannel)
				}
			}
		}(i)
	}
}

func executeGradualIncreaseLoad(ctx context.Context, assignment structs.TaskAssignment, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {
	totalUsers := assignment.LoadTestTestsModel.VirtualUsers
	chunkSize := totalUsers / 10                                                                        // Assuming you want 10 chunks
	delayBetweenChunks := time.Duration(assignment.LoadTestTestsModel.Duration) * time.Millisecond / 10 // 10% of the total duration

	// Initialize a counter to keep track of the current chunk
	currentChunk := 0

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					// Test duration is over.
					return
				default:
					// Determine if user is in the allowed chunk
					if id < (currentChunk+1)*chunkSize {
						// Simulate the virtual user
						simulateVirtualUser(ctx, assignment, id, responseChannel)
					} else {
						// log.Info("User not allowed to make request in this chunk - ", id)
						time.Sleep(500 * time.Millisecond) // If this is removed, the CPU usage will be very high. This could also wait for longer, but it may reduce the accuracy of the test
					}

					// Check if the current chunk has been fully processed
					if id == (currentChunk+1)*chunkSize-1 {
						log.Infof("Chunk %d/%d completed", currentChunk+1, 10)
						// Wait for the delay before allowing the next chunk
						time.Sleep(delayBetweenChunks)
						// Move to the next chunk
						currentChunk++
					}
				}
			}
		}(i)
	}
}

func executeSpikeLoad(ctx context.Context, assignment structs.TaskAssignment, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {
	for i := 0; i < assignment.LoadTestTestsModel.VirtualUsers; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // Random short delay before each spike
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Test duration is over.
					return
				default:
					// Make request
					simulateVirtualUser(ctx, assignment, id, responseChannel)
				}
			}
		}(i)
	}
}

func simulateVirtualUser(ctx context.Context, assignment structs.TaskAssignment, id int, responseChannel chan<- structs.ResponseItem) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	var testPlan []structs.TreeNode
	if err := json.Unmarshal([]byte(assignment.LoadTestPlanModel.TestPlan), &testPlan); err != nil {
		log.Errorf("Failed to parse test plan: %v", err)
		return
	}

	// Execute the test plan nodes
	for _, node := range testPlan {
		select {
		case <-ctx.Done():
			// Test duration is over.
			return
		default:
			executeTreeNode(ctx, node, client, &LastRequestInfo{}, responseChannel)
			log.Infof("Finished executing test plan node (VirtualUser ID: %d)", id)
		}
	}
}

func makeRequest(ctx context.Context, client *http.Client, url string, method string, requestBody []byte, responseChannel chan<- structs.ResponseItem) (lastRequestInfo LastRequestInfo) {
	var req *http.Request
	var err error

	if len(requestBody) > 0 {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		log.Errorf("Failed to create request: %s", err)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: 0}
		return LastRequestInfo{}
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		log.Errorf("Failed to execute request: %s", err)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: elapsed}
		return LastRequestInfo{}
	}
	defer resp.Body.Close()

	responseSize := resp.ContentLength
	responseChannel <- structs.ResponseItem{StatusCode: resp.StatusCode, ResponseTime: elapsed}

	return LastRequestInfo{
		ResponseCode: resp.StatusCode,
		ResponseTime: elapsed,
		ResponseSize: responseSize,
	}
}

type LastRequestInfo struct {
	ResponseCode int
	ResponseTime int64
	ResponseSize int64
}

func executeTreeNode(ctx context.Context, node structs.TreeNode, client *http.Client, lastRequestInfo *LastRequestInfo, responseChannel chan<- structs.ResponseItem) {
	switch node.Type {
	case structs.GetRequest:
		getRequestNode, ok := node.Data.(*structs.GetRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to GetRequestNodeData")
			break
		}
		log.Infof("Making GET request to %s", getRequestNode.URL)

		*lastRequestInfo = makeRequest(ctx, client, getRequestNode.URL, "GET", nil, responseChannel)

	case structs.PostRequest:
		postRequestNode, ok := node.Data.(*structs.PostRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to PostRequestNodeData")
			break
		}
		log.Infof("Making POST request to %s with body: %s", postRequestNode.URL, postRequestNode.Body)

		*lastRequestInfo = makeRequest(ctx, client, postRequestNode.URL, "POST", []byte(postRequestNode.Body), responseChannel)

	case structs.PutRequest:
		putRequestNode, ok := node.Data.(*structs.PutRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to PutRequestNodeData")
			break
		}

		log.Infof("Making PUT request to %s with body: %s", putRequestNode.URL, putRequestNode.Body)

		*lastRequestInfo = makeRequest(ctx, client, putRequestNode.URL, "PUT", []byte(putRequestNode.Body), responseChannel)

	case structs.DeleteRequest:
		deleteRequestNode, ok := node.Data.(*structs.DeleteRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to DeleteRequestNodeData")
			break
		}

		log.Infof("Making DELETE request to %s", deleteRequestNode.URL)

		*lastRequestInfo = makeRequest(ctx, client, deleteRequestNode.URL, "DELETE", nil, responseChannel)

	case structs.IfCondition:
		ifConditionNode, ok := node.Data.(*structs.IfConditionNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to IfConditionNodeData")
			break
		}
		log.Infof("Evaluating if condition at node %s", ifConditionNode.Label)
		log.Infof("Field: %s, Condition: %s, Value: %s", ifConditionNode.Field, ifConditionNode.Condition, ifConditionNode.Value)

		if evaluateField(ifConditionNode, lastRequestInfo) {
			for _, children := range node.Conditions.TrueChildren {
				executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel)
			}
		} else {
			for _, children := range node.Conditions.FalseChildren {
				executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel)
			}
		}
	case structs.DelayNode:
		delayNode, ok := node.Data.(*structs.DelayNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to DelayNodeData")
			break
		}

		switch delayNode.DelayType {
		case structs.Fixed:
			log.Infof("Delaying for %d milliseconds", delayNode.FixedDelay)
			time.Sleep(time.Duration(delayNode.FixedDelay) * time.Millisecond)
		case structs.Random:
			randomDelay := time.Duration(rand.Intn(delayNode.RandomDelay.Max-delayNode.RandomDelay.Min) + delayNode.RandomDelay.Min)
			log.Infof("Delaying for %d milliseconds", randomDelay)
			time.Sleep(randomDelay * time.Millisecond)
		default:
			log.Errorf("Unknown delay type: %s", delayNode.DelayType)
		}

	case structs.StartNode:
		startNode, ok := node.Data.(*structs.StartNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to StartNodeData")
			break
		}
		log.Infof("Starting load test at node %s", startNode.Label)

	case structs.StopNode:
		stopNode, ok := node.Data.(*structs.StopNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to StopNodeData")
			break
		}

		log.Infof("Stopping load test at node %s", stopNode.Label)
		return
	}

	if node.Type != "ifCondition" {
		for _, children := range node.Children {
			executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel)
		}
	}
}

func evaluateField(node *structs.IfConditionNodeData, info *LastRequestInfo) bool {

	switch node.Field {
	case "response_code":
		return evaluateCondition(node, info.ResponseCode)
	case "response_time":
		return evaluateCondition(node, int(info.ResponseTime))
	case "response_size":
		return evaluateCondition(node, int(info.ResponseSize))
	default:
		log.Errorf("Unknown field: %s", node.Field)
		return false
	}
}

func evaluateCondition(node *structs.IfConditionNodeData, value int) bool {

	compareValue, err := strconv.Atoi(node.Value)
	if err != nil {
		log.Errorf("Failed to parse value to int: %s", err)
		return false
	}

	switch node.Condition {
	case "equals":
		return value == compareValue
	case "not_equals":
		return value != compareValue
	case "greater_than":
		return value > compareValue
	case "less_than":
		return value < compareValue
	default:
		log.Errorf("Unknown condition: %s", node.Condition)
		return false
	}
}
