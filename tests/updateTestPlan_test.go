package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/utils"
)

func TestUpdateTestPlan(t *testing.T) {

	loadTestModel, err := CreateLoadTestRequest("Update Test Plan")

	if err != nil {
		t.Fatalf(err.Error())
	}

	testUUID := loadTestModel.UUID

	// Load test plan and react flow plan JSON strings

	testPlan, err := utils.LoadJSONFromFile("../config/testPlans/testPlan_test.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	reactFlow, err := utils.LoadJSONFromFile("../config/testPlans/reactFlowPlan_test.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = UpdateTestPlanRequest(testUUID, testPlan, reactFlow)

	if err != nil {
		t.Fatalf(err.Error())
	}

	loadTestModel, err = GetLoadTestRequest(testUUID)

	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check that the load test was updated (Requires converting)
	err = compareUpdatedValues(loadTestModel, testPlan, reactFlow)
	if err != nil {
		t.Fatalf(err.Error())
	}

	DeleteLoadTestRequest(testUUID)
}

func UpdateTestPlanRequest(testUUID string, testPlan string, reactFlow string) error {
	// Create a new update test plan request
	requestBody, err := createUpdateTestPlanRequest(testPlan, reactFlow)
	if err != nil {
		return err
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return errors.New("Couldn't encode new load test: " + err.Error())
	}

	req, err := http.NewRequest(http.MethodPut, "https://api.load-test.jack.lat/load-tests/"+testUUID+"/plan", bytes.NewBuffer(body))
	if err != nil {
		return errors.New("Couldn't create request: " + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New("Couldn't send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Expected to get status " + strconv.Itoa(http.StatusOK) + " but instead got " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func createUpdateTestPlanRequest(testPlan string, reactFlow string) (*structs.UpdateTestPlanRequest, error) {
	var testPlanObj []structs.TreeNode
	var reactFlowObj structs.ReactFlow

	err := json.Unmarshal([]byte(testPlan), &testPlanObj)
	if err != nil {
		return nil, errors.New("Couldn't decode testPlan: " + err.Error())
	}

	err = json.Unmarshal([]byte(reactFlow), &reactFlowObj)
	if err != nil {
		return nil, errors.New("Couldn't decode reactFlow: " + err.Error())
	}

	// Create a new update test plan request
	updateTestPlanRequest := &structs.UpdateTestPlanRequest{
		TestPlan:  testPlanObj,
		ReactFlow: reactFlowObj,
	}

	return updateTestPlanRequest, nil
}

func compareUpdatedValues(loadTestModel *structs.LoadTestModel, testPlan string, reactFlow string) error {
	// Convert the updated values back to the original structures
	var updatedTestPlan []structs.TreeNode
	var updatedReactFlow structs.ReactFlow

	err := json.Unmarshal([]byte(loadTestModel.TestPlan.TestPlan), &updatedTestPlan)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(loadTestModel.TestPlan.ReactFlowPlan), &updatedReactFlow)
	if err != nil {
		return err
	}

	// Convert the original values back to the original structures
	var originalTestPlan []structs.TreeNode
	var originalReactFlow structs.ReactFlow

	err = json.Unmarshal([]byte(testPlan), &originalTestPlan)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(reactFlow), &originalReactFlow)
	if err != nil {
		return err
	}

	// Compare the updated values with the original values
	if !reflect.DeepEqual(updatedTestPlan, originalTestPlan) {
		return errors.New("Updated testPlan does not match the original testPlan")
	}

	if !reflect.DeepEqual(updatedReactFlow, originalReactFlow) {
		return errors.New("Updated reactFlow does not match the original reactFlow")
	}

	return nil
}
