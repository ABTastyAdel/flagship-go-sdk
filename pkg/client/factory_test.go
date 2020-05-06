package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"

	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/stretchr/testify/assert"
)

func TestCreateClient(t *testing.T) {
	factory := &FlagshipFactory{
		EnvID: testEnvID,
	}

	client, err := factory.CreateClient()

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

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

func TestCreateClientBucketing(t *testing.T) {
	factory := &FlagshipFactory{
		EnvID: testEnvID,
	}

	client, err := factory.CreateClient(WithBucketing())

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

	if len(client.bucketingOptions) != 0 {
		t.Errorf(
			"Bucketing Client default options wrong. Expected default %v, got %v",
			0,
			len(client.bucketingOptions))
	}

	client, err = factory.CreateClient(WithBucketing(bucketing.PollingInterval(2 * time.Second)))

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

	if client.envID != testEnvID {
		t.Error("Wrong env id")
	}

	if client.decisionClient == nil {
		t.Error("decision Bucketing Client has not been initialized")
	}

	bucketing, castOK := client.decisionClient.(*bucketing.Engine)

	if !castOK {
		t.Errorf("decision Bucketing Client has not been initialized correctly")
	}

	pollingInterval := reflect.ValueOf(bucketing).Elem().FieldByName("pollingInterval")
	if pollingInterval.Int() != (2 * time.Second).Nanoseconds() {
		t.Errorf(
			"decision Bucketing Client polling interval wrong. Expected %v, got %v",
			(2 * time.Second).Nanoseconds(),
			pollingInterval.Int())
	}

	if client.batchHitProcessor == nil {
		t.Error("batch hit processor has not been initialized")
	}
}

func TestCreateClientAPIUrl(t *testing.T) {
	factory := &FlagshipFactory{
		EnvID: testEnvID,
	}

	url := "http://google.com"
	client, err := factory.CreateClient(WithDecisionAPI(decision.APIUrl(url)))

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

	apiClient, castOK := client.decisionClient.(*decision.APIClient)
	if !castOK {
		t.Errorf("decision API Client has not been initialized correctly")
	}

	urlClient := reflect.ValueOf(apiClient).Elem().FieldByName("url")
	assert.Equal(t, url, urlClient.String())
}
