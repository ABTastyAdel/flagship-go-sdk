package bucketing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/pkg/utils"
)

const defaultTimeout = 10 * time.Second
const defaultAPIURL = "https://cdn.flagship.io"

var apiLogger = logging.GetLogger("Bucketing API")

// APIClient represents the API client informations
type APIClient struct {
	url         string
	envID       string
	apiKey      string
	timeout     time.Duration
	retries     int
	httpRequest *utils.HTTPRequest
}

// Header element to be sent
type Header struct {
	Name, Value string
}

// APIUrl sets http client base URL
func APIUrl(url string) func(r *APIClient) {
	return func(r *APIClient) {
		r.url = url
	}
}

// APIKey sets http client api key
func APIKey(apiKey string) func(r *APIClient) {
	return func(r *APIClient) {
		r.apiKey = apiKey
	}
}

// Timeout sets http client timeout
func Timeout(timeout time.Duration) func(r *APIClient) {
	return func(r *APIClient) {
		r.timeout = timeout
	}
}

// Retries sets max number of retries for failed calls
func Retries(retries int) func(r *APIClient) {
	return func(r *APIClient) {
		r.retries = retries
	}
}

// NewAPIClient makes Requester with api and parameters. Sets defaults
// api has the base part of request's url, like http://localhost/api/v1
func NewAPIClient(envID string, params ...func(*APIClient)) *APIClient {
	res := APIClient{
		envID:   envID,
		retries: 1,
	}

	headers := []utils.Header{}

	for _, param := range params {
		param(&res)
	}

	if res.apiKey != "" {
		headers = append(headers, utils.Header{Name: "x-api-key", Value: res.apiKey})
	}

	if res.url == "" {
		res.url = defaultAPIURL
	}

	res.httpRequest = utils.NewHTTPRequest(res.url, utils.HTTPOptions{
		Timeout: res.timeout,
		Headers: headers,
	})

	return &res
}

// GetConfiguration gets an environment configuration from API
func (r APIClient) GetConfiguration() (*Configuration, error) {
	path := fmt.Sprintf("/%s/bucketing.json", r.envID)

	response, _, code, err := r.httpRequest.Do(path, "GET", nil)

	if err != nil {
		return nil, err
	}

	if code != 200 && code != 304 {
		return nil, fmt.Errorf("Error when calling Bucketing API : %v", err)
	}

	resp := &Configuration{}
	err = json.Unmarshal(response, &resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
