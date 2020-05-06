package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

const defaultTimeout = 2 * time.Second

var apiLogger = logging.GetLogger("HTTP Request")

// HTTPRequest represents the HTTPRequest infos
type HTTPRequest struct {
	baseURL string
	client  http.Client
	HTTPOptions
}

// HTTPOptions represents the options for the HTTPRequest object
type HTTPOptions struct {
	Retries int
	Timeout time.Duration
	Headers []Header
}

// Header element to be sent
type Header struct {
	Name, Value string
}

// NewHTTPRequest creates an HTTP requester object
func NewHTTPRequest(baseURL string, options HTTPOptions) *HTTPRequest {
	headers := []Header{{"Content-Type", "application/json"}, {"Accept", "application/json"}}
	retries := 1
	timeout := defaultTimeout

	if options.Retries > 1 {
		retries = options.Retries
	}

	if options.Timeout != 0 {
		timeout = options.Timeout
	}

	if options.Headers != nil {
		for _, h := range options.Headers {
			headers = append(headers, h)
		}
	}

	client := http.Client{Timeout: timeout}

	res := HTTPRequest{
		baseURL: baseURL,
		HTTPOptions: HTTPOptions{
			Timeout: timeout,
			Headers: headers,
			Retries: retries,
		},
		client: client,
	}

	return &res
}

// Do executes request and returns response body for requested url
func (r HTTPRequest) Do(path, method string, body io.Reader) (response []byte, responseHeaders http.Header, code int, err error) {
	single := func(request *http.Request) (response []byte, responseHeaders http.Header, code int, e error) {
		resp, doErr := r.client.Do(request)
		if doErr != nil {
			apiLogger.Error(fmt.Sprintf("failed to send request %v", request), e)
			return nil, http.Header{}, 0, doErr
		}
		defer func() {
			if e := resp.Body.Close(); e != nil {
				apiLogger.Warning(fmt.Sprintf("can't close body for %s request, %s", request.URL, e))
			}
		}()

		if response, err = ioutil.ReadAll(resp.Body); err != nil {
			apiLogger.Error("failed to read body", err)
			return nil, resp.Header, resp.StatusCode, err
		}

		if resp.StatusCode >= http.StatusBadRequest {
			apiLogger.Warning(fmt.Sprintf("error status code=%d", resp.StatusCode))
			return response, resp.Header, resp.StatusCode, errors.New(resp.Status)
		}

		return response, resp.Header, resp.StatusCode, nil
	}

	url := fmt.Sprintf("%s%s", r.baseURL, path)
	apiLogger.Debug(fmt.Sprintf("requesting %s", url))

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		apiLogger.Error(fmt.Sprintf("failed to make request %s", url), err)
		return nil, nil, 0, err
	}

	for _, h := range r.Headers {
		req.Header.Add(h.Name, h.Value)
	}

	for i := 0; i < r.Retries; i++ {
		if response, responseHeaders, code, err = single(req); err == nil {
			triedMsg := ""
			if i > 0 {
				triedMsg = fmt.Sprintf(", tried %d time(s)", i+1)
			}
			apiLogger.Debug(fmt.Sprintf("completed %s%s", url, triedMsg))
			return response, responseHeaders, code, err
		}
		apiLogger.Debug(fmt.Sprintf("failed %s with %v", url, err))

		if i != r.Retries {
			delay := time.Duration(100) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return response, responseHeaders, code, err
}
