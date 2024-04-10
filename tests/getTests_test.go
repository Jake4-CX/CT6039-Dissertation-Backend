package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
)

func TestGetTests(t *testing.T) {
	resp, err := http.Get("https://api.load-test.jack.lat/load-tests")
	if err != nil {
		t.Fatalf("Couldn't create request: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, resp.StatusCode)
	}

	var loadTests []structs.LoadTestModel
	err = json.NewDecoder(resp.Body).Decode(&loadTests)
	if err != nil {
		t.Fatalf("Couldn't decode response body: %s\n", err)
	}

	// Check that loadTests is not empty
	if len(loadTests) == 0 {
		t.Fatalf("Expected to get at least one load test but instead got none\n")
	}
}

func GetLoadTestRequest(testUUID string) (*structs.LoadTestModel, error) {
	// Send a GET request to the /load-tests/:id endpoint
	req, err := http.NewRequest(http.MethodGet, "https://api.load-test.jack.lat/load-tests/"+testUUID, nil)
	if err != nil {
		return nil, errors.New("Couldn't create request: " + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Couldn't send request: " + err.Error())
	}
	defer resp.Body.Close()

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Expected to get status " + string(rune(http.StatusOK)) + " but instead got " + string(rune(resp.StatusCode)))
	}

	var response struct {
		Test        structs.LoadTestModel                          `json:"test"`
		TestMetrics map[uint]*map[int64][]structs.ResponseFragment `json:"testMetrics"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, errors.New("Couldn't decode response body: " + err.Error())
	}

	return &response.Test, nil
}
