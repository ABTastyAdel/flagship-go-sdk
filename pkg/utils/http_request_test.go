package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPRequest(t *testing.T) {
	url := "http://google.fr"
	r := NewHTTPRequest(url, HTTPOptions{})

	if r == nil {
		t.Error("HTTP Request should not be empty")
	}

	if r.baseURL != url {
		t.Errorf("HTTP Request url incorrect. Should be default %v, got %v", url, r.baseURL)
	}

	if r.Retries != 1 {
		t.Errorf("HTTP Request retries incorrect. Should be default 1, got %v", r.Retries)
	}

	if r.Timeout != defaultTimeout {
		t.Errorf("HTTP Request timeout incorrect. Should be default %v, got %v", defaultTimeout, r.Timeout)
	}

	if r.client.Timeout != defaultTimeout {
		t.Errorf("HTTP Request client timeout incorrect. Should be default %v, got %v", defaultTimeout, r.client.Timeout)
	}

	if len(r.Headers) != 2 {
		t.Errorf("HTTP Request headers incorrect. Should be default %v, got %v", 2, len(r.Headers))
	}
}

func TestNewHTTPRequestOptions(t *testing.T) {
	url := "http://google.fr"
	r := NewHTTPRequest(url, HTTPOptions{
		Retries: 2,
		Timeout: 10,
		Headers: []Header{
			Header{Name: "test", Value: "value"},
		},
	})

	if r.Retries != 2 {
		t.Errorf("HTTP Request retries incorrect. Should be 2, got %v", r.Retries)
	}

	if r.Timeout != 10 {
		t.Errorf("HTTP Request timeout incorrect. Should be default %v, got %v", 10, r.Timeout)
	}

	if r.client.Timeout != 10 {
		t.Errorf("HTTP Request client timeout incorrect. Should be default %v, got %v", 10, r.client.Timeout)
	}

	if len(r.Headers) != 3 {
		t.Errorf("HTTP Request headers incorrect. Should be default %v, got %v", 3, len(r.Headers))
	}
}

func TestDo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, "ok")
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	httpreq := NewHTTPRequest(ts.URL, HTTPOptions{})
	resp, _, _, _ := httpreq.Do("/good", "GET", nil)
	if "ok\n" != string(resp) {
		t.Errorf("Expected response %v, got %v", "ok\n", resp)
	}

	_, _, code, err := httpreq.Do("/bad", "GET", nil)
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
	if code != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, code)
	}
}

func TestPost(t *testing.T) {
	type body struct {
		testString string
		testInt    int
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, "ok")
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	bodyObj := body{"one", 1}
	b, _ := json.Marshal(bodyObj)

	httpreq := NewHTTPRequest(ts.URL, HTTPOptions{})
	resp, _, _, _ := httpreq.Do("/good", "POST", bytes.NewBuffer(b))
	if "ok\n" != string(resp) {
		t.Errorf("Expected response %v, got %v", "ok\n", resp)
	}

	_, _, code, err := httpreq.Do("/bad", "POST", bytes.NewBuffer(b))
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
	if code != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, code)
	}
}

func TestGetRetry(t *testing.T) {
	nbCalls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/retry" {
			nbCalls++
			if nbCalls >= 5 {
				fmt.Fprintln(w, "ok")
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	httpreq := NewHTTPRequest(ts.URL, HTTPOptions{
		Retries: 10,
	})

	resp, _, _, _ := httpreq.Do("/retry", "GET", nil)
	if "ok\n" != string(resp) {
		t.Errorf("Expected response %v, got %v", "ok\n", resp)
	}
	if 5 != nbCalls {
		t.Errorf("Expected %v http calls, got %v", 5, nbCalls)
	}

	httpreq = NewHTTPRequest(ts.URL, HTTPOptions{
		Retries: 3,
	})
	nbCalls = 0
	_, _, code, err := httpreq.Do("/retry", "GET", nil)
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
	if code != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, code)
	}
	if 3 != nbCalls {
		t.Errorf("Expected %v http calls, got %v", 3, nbCalls)
	}
}

func TestFailCall(t *testing.T) {
	httpreq := NewHTTPRequest("wrong_url", HTTPOptions{})

	_, _, _, err := httpreq.Do("/", "GET", nil)
	assert.NotNil(t, err)
}
