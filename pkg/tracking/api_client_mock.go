package tracking

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// MockAPIClient represents a fake API client informations
type MockAPIClient struct {
	envID      string
	shouldFail bool
}

// NewMockAPIClient makes Requester with api and parameters. Sets defaults
// api has the base part of request's url, like http://localhost/api/v1
func NewMockAPIClient(envID string, shouldFail bool) *MockAPIClient {
	res := MockAPIClient{
		shouldFail: shouldFail,
	}

	return &res
}

// SendHit sends a tracking hit to the Data Collect API
func (r MockAPIClient) sendInternalHit(hit HitInterface) error {
	errs := hit.validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Hit validation error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid hit : %s", strings.Join(errorStrings, ", "))
	}
	hit.computeQueueTime()

	json, err := json.Marshal(hit)

	log.Printf("Sending hit : %v", string(json))

	if r.shouldFail {
		return errors.New("Mock fail send hit error")
	}

	return err
}

// ActivateCampaign activate a campaign / variation id to the Decision API
func (r MockAPIClient) ActivateCampaign(request ActivationHit) error {
	request.EnvironmentID = r.envID

	if r.shouldFail {
		return errors.New("Mock fail activate error")
	}

	return nil
}
