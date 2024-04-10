package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
)

func TestCreateLoadTest(t *testing.T) {
    testName := "Test Load Test"

    loadTestRequest, err := CreateLoadTestRequest(testName)

    if err != nil {
        t.Fatalf(err.Error())
    }

    DeleteLoadTestRequest(loadTestRequest.UUID) // Delete test - (not tested here)
}

func CreateLoadTestRequest(testName string) (*structs.LoadTestModel, error) {
	// Create a new load test
	newLoadTest := struct {
		Name string `json:"name"`
	}{
		Name: testName,
	}
	body, err := json.Marshal(newLoadTest)
	if err != nil {
		return nil, errors.New("Couldn't encode new load test: " + err.Error())
	}

	resp, err := http.Post("https://api.load-test.jack.lat/load-tests", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New("Couldn't create request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Expected to get status " + string(rune(http.StatusOK)) + " but instead got " + string(rune(resp.StatusCode)))
	}

	var response struct {
		Message string                `json:"message"`
		Data    structs.LoadTestModel `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, errors.New("Couldn't decode response body: " + err.Error())
	}

	// Check that the load test was created
	if response.Data.Name != newLoadTest.Name {
		return nil, errors.New("Expected to get load test with name '" + newLoadTest.Name + "' but instead got '" + response.Data.Name + "'")
	}

	return &response.Data, nil
}
