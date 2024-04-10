package tests

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestDeleteLoadTest(t *testing.T) {
    loadTestModel, err := CreateLoadTestRequest("Delete Load Test")

		if err != nil {
			t.Fatalf(err.Error())
		}

		err = DeleteLoadTestRequest(loadTestModel.UUID)

		if err != nil {
			t.Fatalf(err.Error())
		}

		_, err = GetLoadTestRequest(loadTestModel.UUID)

		if err == nil {
			t.Fatalf("Expected to get an error but instead got none - load test was not deleted")
		}
}

func DeleteLoadTestRequest(testUUID string) (error) {
		// Send DELETE request to /load-tests/:id endpoint
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.load-test.jack.lat/load-tests/%s", testUUID), nil)
		if err != nil {
				return errors.New("Couldn't create request: " + err.Error())
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
				return errors.New("Couldn't send request: " + err.Error())
		}
		defer resp.Body.Close()

		// Check the status code of the response
		if resp.StatusCode != http.StatusOK {
				return errors.New("Expected to get status " + string(rune(http.StatusOK)) + " but instead got " + string(rune(resp.StatusCode)))
		}

		return nil
}