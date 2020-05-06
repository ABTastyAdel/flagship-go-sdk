package decision

import (
	"testing"
)

var testEnvID = "env_id_test"
var realEnvID = "blvo2kijq6pg023l8edg"

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testEnvID)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultAPIURL {
		t.Error("Api url should be set to default")
	}
}

func TestNewAPIClientParams(t *testing.T) {
	client := NewAPIClient(
		testEnvID,
		APIUrl("https://google.com"),
		APIKey("api_key"),
		Timeout(10),
		Retries(12))

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != "https://google.com" {
		t.Error("Api url should be set to default")
	}

	if client.apiKey != "api_key" {
		t.Errorf("Wrong api key. Expected %v, got %v", "api_key", client.apiKey)
	}

	if client.retries != 12 {
		t.Errorf("Wrong retries. Expected %v, got %v", 12, client.retries)
	}

	if client.httpRequest.Timeout != 10 {
		t.Errorf("Wrong timeout. Expected %v, got %v", 10, client.httpRequest.Timeout)
	}

	if len(client.httpRequest.Headers) != 3 {
		t.Errorf("Wrong headers. Expected %v, got %v", 3, len(client.httpRequest.Headers))
	}
}

func TestGetModifications(t *testing.T) {
	client := NewAPIClient(testEnvID)
	_, err := client.GetModifications("test_vid", nil)

	if err == nil {
		t.Error("Expected error for unknown env id")
	}

	client = NewAPIClient(realEnvID)
	resp, err := client.GetModifications("test_vid", nil)

	if err != nil {
		t.Errorf("Unexpected error for correct env id : %v", err)
	}

	if resp == nil {
		t.Errorf("Expected not nil response for correct env id")
	}
}
