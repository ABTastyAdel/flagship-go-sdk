package tracking

import (
	"testing"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testEnvID)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.urlDecision != defaultAPIURLDecision {
		t.Error("Api url should be set to default")
	}

	if client.urlTracking != defaultAPIURLTracking {
		t.Error("Api url should be set to default")
	}
}

func TestSendInternalHit(t *testing.T) {
	client := NewAPIClient(testEnvID)
	err := client.sendInternalHit(nil)

	if err == nil {
		t.Error("Empty hit should return and err")
	}

	event := &EventHit{}
	event.setBaseInfos(testEnvID, testVisitorID)

	err = client.sendInternalHit(event)

	if err == nil {
		t.Error("Invalid event hit should return error")
	}

	event.Action = "test_action"
	err = client.sendInternalHit(event)

	if err != nil {
		t.Error("Right hit should not return and err")
	}
}

func TestNewAPIClientParams(t *testing.T) {
	client := NewAPIClient(testEnvID, DecisionAPIKey("youpi"), DecisionTimeout(10))

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.apiKey != "youpi" {
		t.Errorf("Wrong api key. Expected %v, got %v", "youpi", client.apiKey)
	}

	if client.httpRequestDecision.Timeout != 10 {
		t.Errorf("Wrong timeout. Expected %v, got %v", 10, client.httpRequestDecision.Timeout)
	}

	if len(client.httpRequestDecision.Headers) != 3 {
		t.Errorf("Wrong headers. Expected %v, got %v", 3, len(client.httpRequestDecision.Headers))
	}
}

func TestActivate(t *testing.T) {
	client := NewAPIClient(testEnvID)
	err := client.ActivateCampaign(ActivationHit{})

	if err == nil {
		t.Errorf("Expected error for empty request")
	}

	err = client.ActivateCampaign(ActivationHit{
		EnvironmentID:    testEnvID,
		VisitorID:        "test_vid",
		VariationGroupID: "vgid",
		VariationID:      "vid",
	})

	if err != nil {
		t.Errorf("Did not expect error for correct activation request. Got %v", err)
	}
}
