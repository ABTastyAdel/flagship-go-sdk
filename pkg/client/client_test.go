package client

import (
	"errors"
	"testing"

	"github.com/abtasty/flagship-go-sdk/pkg/tracking"
)

var testEnvID = "test_env_id"
var vID = "test_visitor_id"

func createClient() *FlagshipClient {
	return &FlagshipClient{
		envID: testEnvID,
	}
}

func TestInit(t *testing.T) {
	client := createClient()
	client.init()

	if client.envID != testEnvID {
		t.Error("Wrong env id")
	}

	if client.decisionClient == nil {
		t.Error("decision API Client has not been initialized")
	}

	if client.batchHitProcessor == nil {
		t.Error("batch hit processor has not been initialized")
	}
}

func TestCreateVisitor(t *testing.T) {
	client := &FlagshipClient{
		envID: testEnvID,
	}
	client.init()

	context := map[string]interface{}{}
	context["test_string"] = "123"
	context["test_number"] = 36.5
	context["test_bool"] = true
	context["test_int"] = 4
	context["test_wrong"] = errors.New("wrong type")

	_, err := client.NewVisitor("", nil)

	if err != nil {
		t.Error("Empty visitor ID should raise an error")
	}

	_, err = client.NewVisitor(vID, context)

	if err == nil {
		t.Error("Visitor with wrong context variable should raise an error")
	}

	_, conv64Ok := context["test_int"].(float64)
	if !conv64Ok {
		t.Errorf("Integer context key has not been converted. Got %v", context["test_int"])
	}

	delete(context, "test_wrong")

	visitor, err := client.NewVisitor(vID, context)
	if err != nil {
		t.Errorf("Visitor creation failed. error : %v", err)
	}

	if visitor == nil {
		t.Error("Visitor creation failed. Visitor is null")
	}

	if visitor.ID != vID {
		t.Error("Visitor creation failed. Visitor id is not set")
	}

	for key, val := range context {
		valV, exists := visitor.Context[key]
		if !exists {
			t.Errorf("Visitor creation failed. Visitor context key %s is not set", key)
		}
		if val != valV {
			t.Errorf("Visitor creation failed. Visitor context key %s value %v is wrong. Should be %v", key, valV, val)
		}
	}
}

func TestSendHitClient(t *testing.T) {
	client := &FlagshipClient{
		envID: testEnvID,
	}
	client.init()

	err := client.SendHit(vID, &tracking.EventHit{})

	if err == nil {
		t.Errorf("Expected error as hit is malformed.")
	}

	err = client.SendHit(vID, &tracking.EventHit{
		Action: "test_action",
	})
	if err != nil {
		t.Errorf("Did not expect error as hit is correct. Got %v", err)
	}
}
