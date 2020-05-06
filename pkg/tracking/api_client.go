package tracking

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/pkg/utils"
)

const defaultTimeout = 2 * time.Second
const defaultAPIURLTracking = "https://ariane.abtasty.com"
const defaultAPIURLDecision = "https://decision-api.flagship.io"

var apiLogger = logging.GetLogger("DataCollect API")

// APIClient represents the API client informations
type APIClient struct {
	urlTracking         string
	urlDecision         string
	envID               string
	decisionTimeout     time.Duration
	apiKey              string
	httpRequestTracking *utils.HTTPRequest
	httpRequestDecision *utils.HTTPRequest
}

// Header element to be sent
type Header struct {
	Name, Value string
}

// DecisionAPIKey sets http client api key for the decision API calls
func DecisionAPIKey(apiKey string) func(r *APIClient) {
	return func(r *APIClient) {
		r.apiKey = apiKey
	}
}

// DecisionTimeout sets http client timeout for decision calls
func DecisionTimeout(timeout time.Duration) func(r *APIClient) {
	return func(r *APIClient) {
		r.decisionTimeout = timeout
	}
}

// NewAPIClient makes Requester with api and parameters. Sets defaults
// api has the base part of request's url, like http://localhost/api/v1
func NewAPIClient(envID string, params ...func(r *APIClient)) *APIClient {
	res := APIClient{
		envID: envID,
	}

	headers := []utils.Header{}

	for _, param := range params {
		param(&res)
	}

	if res.urlTracking == "" {
		res.urlTracking = defaultAPIURLTracking
	}

	if res.urlDecision == "" {
		res.urlDecision = defaultAPIURLDecision
	}

	if res.apiKey != "" {
		headers = append(headers, utils.Header{Name: "x-api-key", Value: res.apiKey})
	}

	httpRequestDecision := utils.NewHTTPRequest(res.urlDecision, utils.HTTPOptions{
		Timeout: res.decisionTimeout,
		Headers: headers,
	})
	httpRequestTracking := utils.NewHTTPRequest(res.urlTracking, utils.HTTPOptions{})

	res.httpRequestDecision = httpRequestDecision
	res.httpRequestTracking = httpRequestTracking

	return &res
}

// sendInternalHit sends a tracking hit to the Data Collect API
func (r APIClient) sendInternalHit(hit HitInterface) error {
	if hit == nil {
		err := errors.New("Hit should not be empty")
		apiLogger.Error(err.Error(), err)
		return err
	}

	errs := hit.validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Hit validation error : %v", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return errors.New("Hit validation failed")
	}
	hit.computeQueueTime()

	b, err := json.Marshal(hit)

	if err != nil {
		return err
	}

	apiLogger.Info(fmt.Sprintf("Sending hit : %v", string(b)))
	_, _, code, err := r.httpRequestTracking.Do("", "POST", bytes.NewBuffer(b))

	if err != nil {
		return err
	}

	if code != 200 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}

// ActivateCampaign activate a campaign / variation id to the Decision API
func (r APIClient) ActivateCampaign(request ActivationHit) error {
	request.EnvironmentID = r.envID

	errs := request.validate()

	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Activate hit validation error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid activation hit : %s", strings.Join(errorStrings, ", "))
	}

	b, err := json.Marshal(request)

	if err != nil {
		return err
	}
	_, _, code, err := r.httpRequestDecision.Do("/v1/activate", "POST", bytes.NewBuffer(b))

	if err != nil {
		return err
	}

	if code != 204 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}
