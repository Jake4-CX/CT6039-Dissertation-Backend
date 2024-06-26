package services

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/worker/state"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
)

var activeGoroutines int64
var activeGoroutineIDs []int

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
	responseChannel := make(chan structs.ResponseItem, assignment.LoadTestTestsModel.VirtualUsers+10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(assignment.LoadTestTestsModel.Duration)*time.Millisecond)
	defer cancel()

	state.LoadTestCancellers.Store(assignment.LoadTestTestsModel.ID, cancel)

	// Start reporting metrics periodically
	go ReportMetricsPeriodically(ctx, assignment.AssignedWorkerID, responseChannel, assignment.LoadTestTestsModel.Duration, assignment.LoadTestTestsModel.ID)

	var testPlan []structs.TreeNode
	if err := json.Unmarshal([]byte(assignment.LoadTestPlanModel.TestPlan), &testPlan); err != nil {
		log.Errorf("Failed to parse test plan: %v", err)
		return
	}

	testStartTime := time.Now()

	switch assignment.LoadTestTestsModel.LoadTestType {
	case structs.Load, structs.Soak:
		log.Info("Executing constant load test")
		executeConstantLoad(ctx, assignment, testPlan, &wg, responseChannel)

	case structs.Stress:
		log.Info("Executing stress test")
		executeGradualIncreaseLoad(ctx, assignment, testPlan, &wg, responseChannel)

	case structs.Spike:
		log.Info("Executing spike test")
		executeSpikeLoad(ctx, assignment, testPlan, &wg, responseChannel)

	default:
		log.Errorf("Unknown load test type: %s", assignment.LoadTestTestsModel.LoadTestType)
	}

	// go func() {
	// 	// loop until the context is done
	// 	for {
	// 		if atomic.LoadInt64(&activeGoroutines) == 0 {
	// 			break
	// 		}
	// 		log.Infof("Active goroutines: %d", atomic.LoadInt64(&activeGoroutines))
	// 		log.Infof("Active goroutine IDs: %v", activeGoroutineIDs)
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	wg.Wait()

	log.Infof("All goroutines finished. Active goroutines: %d", atomic.LoadInt64(&activeGoroutines))

	close(responseChannel)

	log.Info("Waiting for metrics to be reported")

	testDuration := time.Since(testStartTime)

	log.Infof("Load test completed in %s.", testDuration)
}

// Put Load Test Type Executor functions here
func executeConstantLoad(ctx context.Context, assignment structs.TaskAssignment, testPlan []structs.TreeNode, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {

	// make an array of goroutines Ids

	for i := 0; i < assignment.LoadTestTestsModel.VirtualUsers; i++ {
		wg.Add(1)
		atomic.AddInt64(&activeGoroutines, 1)
		activeGoroutineIDs = append(activeGoroutineIDs, i)
		go func(id int) {
			defer wg.Done()
			defer atomic.AddInt64(&activeGoroutines, -1)
			defer func() {
				// Remove the goroutine ID from the active goroutine IDs slice
				for i, v := range activeGoroutineIDs {
					if v == id {
						activeGoroutineIDs = append(activeGoroutineIDs[:i], activeGoroutineIDs[i+1:]...)
						break
					}
				}
			}()
			for {
				select {
				case <-ctx.Done():
					log.Infof("(Virtual User: %d) Test duration is over", id)
					// Test duration is over.
					return
				default:
					// Make request
					simulateVirtualUser(ctx, testPlan, id, responseChannel)
				}
			}
		}(i)
	}
}

func executeGradualIncreaseLoad(ctx context.Context, assignment structs.TaskAssignment, testPlan []structs.TreeNode, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {
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
						simulateVirtualUser(ctx, testPlan, id, responseChannel)
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

func executeSpikeLoad(ctx context.Context, assignment structs.TaskAssignment, testPlan []structs.TreeNode, wg *sync.WaitGroup, responseChannel chan<- structs.ResponseItem) {
	totalUsers := assignment.LoadTestTestsModel.VirtualUsers
	peakUsers := totalUsers          // 100% of total users
	lowUsers := totalUsers / 5       // 20% of total users
	cycleDuration := time.Second * 4 // 4 seconds per cycle

	// Initialize a counter to keep track of the current cycle
	currentCycle := 0

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
					var usersForCycle int
					if !(currentCycle%2 == 0) {
						// Low users
						usersForCycle = lowUsers
						if id < lowUsers {
							// Simulate the virtual user
							simulateVirtualUser(ctx, testPlan, id, responseChannel)
						} else {
							time.Sleep(500 * time.Millisecond) // If this is removed, the CPU usage will be very high. This could also wait for longer, but it may reduce the accuracy of the test
						}
					} else {
						// Peak users
						usersForCycle = peakUsers
						simulateVirtualUser(ctx, testPlan, id, responseChannel)
					}

					if id == usersForCycle-1 {
						log.Infof("Chunk %d/%d completed", usersForCycle, totalUsers)
						time.Sleep(cycleDuration)
						currentCycle++
					}
				}
			}

		}(i)
	}
}

func simulateVirtualUser(ctx context.Context, testPlan []structs.TreeNode, id int, responseChannel chan<- structs.ResponseItem) {

	client := &http.Client{
		Timeout: 3000 * time.Millisecond,
	}
	// Execute the test plan nodes
	for _, node := range testPlan {
		select {
		case <-ctx.Done():
			// Test duration is over.
			return
		default:
			executeTreeNode(ctx, node, client, &LastRequestInfo{}, responseChannel, id)
			log.Debugf("Finished executing test plan node (VirtualUser ID: %d)", id)
		}
	}
}

func makeRequest(ctx context.Context, client *http.Client, url string, method string, requestBody []byte, responseChannel chan<- structs.ResponseItem, id int) (lastRequestInfo LastRequestInfo) {
	var req *http.Request
	var err error

	if len(requestBody) > 0 {
		// log.Infof("VIRTUALUSER: %d REQUEST:A-POST/PUT", id)
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
		// log.Infof("VIRTUALUSER: %d REQUEST:B-POST/PUT", id)
	} else {
		// log.Infof("VIRTUALUSER: %d REQUEST:A-REGULAR", id)
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		// log.Infof("VIRTUALUSER: %d REQUEST:B-REGULAR", id)
	}

	if err != nil {
		log.Errorf("Failed to create request: %s", err)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: 0}
		return LastRequestInfo{
			ResponseCode: 0,
			ResponseTime: 0,
			ResponseSize: 0,
		}
	}

	// log.Infof("VIRTUALUSER: %d REQUEST:C", id)
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	// log.Infof("VIRTUALUSER: %d REQUEST:D", id)

	if err != nil {
		log.Infof("(Virtual User: %d) Failed to execute request: %s", id, err)
		// log.Infof("VIRTUALUSER: %d REQUEST:E1", id)
		responseChannel <- structs.ResponseItem{StatusCode: 0, ResponseTime: elapsed}
		// log.Infof("VIRTUALUSER: %d REQUEST:E2", id)
		return LastRequestInfo{
			ResponseCode: 0,
			ResponseTime: elapsed,
			ResponseSize: 0,
		}
	}
	defer resp.Body.Close()

	// log.Infof("VIRTUALUSER: %d REQUEST:F", id)

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

func executeTreeNode(ctx context.Context, node structs.TreeNode, client *http.Client, lastRequestInfo *LastRequestInfo, responseChannel chan<- structs.ResponseItem, id int) {
	// log.Infof("VIRTUALUSER: %d EXECUTING NODE: %s", id, node.Type)

	switch node.Type {
	case structs.GetRequest:
		getRequestNode, ok := node.Data.(*structs.GetRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to GetRequestNodeData")
			break
		}
		log.Debugf("Making GET request to %s", getRequestNode.URL)

		*lastRequestInfo = makeRequest(ctx, client, getRequestNode.URL, "GET", nil, responseChannel, id)

	case structs.PostRequest:
		postRequestNode, ok := node.Data.(*structs.PostRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to PostRequestNodeData")
			break
		}
		log.Debugf("Making POST request to %s with body: %s", postRequestNode.URL, postRequestNode.Body)

		*lastRequestInfo = makeRequest(ctx, client, postRequestNode.URL, "POST", []byte(postRequestNode.Body), responseChannel, id)

	case structs.PutRequest:
		putRequestNode, ok := node.Data.(*structs.PutRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to PutRequestNodeData")
			break
		}

		log.Debugf("Making PUT request to %s with body: %s", putRequestNode.URL, putRequestNode.Body)

		*lastRequestInfo = makeRequest(ctx, client, putRequestNode.URL, "PUT", []byte(putRequestNode.Body), responseChannel, id)

	case structs.DeleteRequest:
		deleteRequestNode, ok := node.Data.(*structs.DeleteRequestNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to DeleteRequestNodeData")
			break
		}

		log.Debugf("Making DELETE request to %s", deleteRequestNode.URL)

		*lastRequestInfo = makeRequest(ctx, client, deleteRequestNode.URL, "DELETE", nil, responseChannel, id)

	case structs.IfCondition:
		ifConditionNode, ok := node.Data.(*structs.IfConditionNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to IfConditionNodeData")
			break
		}
		log.Debugf("Evaluating if condition at node %s", ifConditionNode.Label)
		log.Debugf("Field: %s, Condition: %s, Value: %s", ifConditionNode.Field, ifConditionNode.Condition, ifConditionNode.Value)

		if evaluateField(ifConditionNode, lastRequestInfo) {
			for _, children := range node.Conditions.TrueChildren {
				executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel, id)

				if ctx.Err() != nil {
					return
				}
			}
		} else {
			for _, children := range node.Conditions.FalseChildren {
				executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel, id)

				if ctx.Err() != nil {
					return
				}
			}
		}
	case structs.DelayNode:
		delayNode, ok := node.Data.(*structs.DelayNodeData)
		if !ok {
			log.Debugf("Failed to cast node data to DelayNodeData")
			break
		}

		switch delayNode.DelayType {
		case structs.Fixed:
			log.Debugf("Delaying for %d milliseconds", delayNode.FixedDelay)
			time.Sleep(time.Duration(delayNode.FixedDelay) * time.Millisecond)
		case structs.Random:
			randomDelay := time.Duration(rand.Intn(delayNode.RandomDelay.Max-delayNode.RandomDelay.Min) + delayNode.RandomDelay.Min)
			log.Debugf("Delaying for %d milliseconds", randomDelay)
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
		log.Debugf("Starting load test at node %s", startNode.Label)

	case structs.StopNode:
		stopNode, ok := node.Data.(*structs.StopNodeData)
		if !ok {
			log.Errorf("Failed to cast node data to StopNodeData")
			break
		}

		log.Debugf("Stopping load test at node %s", stopNode.Label)
		return
	}

	if node.Type != "ifCondition" && ctx.Err() == nil {
		for _, children := range node.Children {
			executeTreeNode(ctx, children, client, lastRequestInfo, responseChannel, id)

			if ctx.Err() != nil {
				return
			}
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
