package bucketing

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/utils"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/stretchr/testify/assert"
)

var testVID = "test_vid"
var testContext = map[string]interface{}{}

func TestNewEngine(t *testing.T) {
	eg := utils.NewExecGroup(context.Background())
	engine, err := NewEngine(testEnvID, eg)

	if err == nil {
		t.Error("Bucketing engine creation should return an error for incorrect envID")
	}

	if engine == nil {
		t.Error("Bucketing engine should not be nil")
	}

	if engine.envID != testEnvID {
		t.Errorf("Bucketing engine env ID incorrect. Expected %v, got %v", testEnvID, engine.envID)
	}

	url := "http://google.fr"
	engine, err = NewEngine(testEnvID, eg, APIOptions(APIUrl(url)))

	if err == nil {
		t.Error("Bucketing engine creation should return an error for incorrect url")
	}

	if engine == nil {
		t.Error("Bucketing engine should not be nil")
	}

	apiClient, castOK := engine.apiClient.(*APIClient)
	if !castOK {
		t.Errorf("bucketing API Client has not been initialized correctly")
	}

	urlClient := reflect.ValueOf(apiClient).Elem().FieldByName("url")
	assert.Equal(t, url, urlClient.String())
}

func TestLoad(t *testing.T) {
	eg := utils.NewExecGroup(context.Background())
	engine, _ := NewEngine(testEnvID, eg)
	err := engine.Load()

	if err == nil {
		t.Error("Expected error for incorrect env ID")
	}

	engine, _ = NewEngine(realEnvID, eg)
	err = engine.Load()

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
}

func TestGetModifications(t *testing.T) {
	eg := utils.NewExecGroup(context.Background())
	engine, _ := NewEngine(testEnvID, eg)

	modifs, err := engine.GetModifications(testVID, testContext)

	if err == nil {
		t.Errorf("Expected error for test env ID")
	}

	if modifs != nil {
		t.Errorf("Unexpected modifs for test env ID. Got %v", modifs)
	}

	engine, _ = NewEngine(realEnvID, eg)

	_, err = engine.GetModifications(testVID, testContext)

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
}

func TestPanic(t *testing.T) {
	eg := utils.NewExecGroup(context.Background())
	engine, _ := NewEngine(testEnvID, eg)

	config := &Configuration{
		Campaigns: []*Campaign{{
			ID: "test_cid",
			VariationGroups: []*VariationGroup{{
				ID: "test_vgid",
				Targeting: TargetingWrapper{
					TargetingGroups: []*TargetingGroup{{
						Targetings: []*Targeting{{
							Operator: EQUALS,
							Key:      "test",
							Value:    true,
						}},
					}},
				},
				Variations: []*Variation{{
					ID:         "1",
					Allocation: 100,
					Modifications: decision.APIClientModification{
						Type:  "FLAG",
						Value: map[string]interface{}{"test": true},
					},
				}},
			}},
		}},
	}

	engine.apiClient = NewAPIClientMock(testEnvID, config, 200)

	modifs, err := engine.GetModifications(testVID, map[string]interface{}{"test": true})

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
	assert.Equal(t, 1, len(modifs.Campaigns))

	// Setting panic
	config.Panic = true
	engine.apiClient = NewAPIClientMock(testEnvID, config, 200)

	modifs, err = engine.GetModifications(testVID, map[string]interface{}{"test": true})

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
	assert.Equal(t, 0, len(modifs.Campaigns))
}

func TestPollingPanic(t *testing.T) {
	eg := utils.NewExecGroup(context.Background())
	engine, _ := NewEngine(testEnvID, eg, PollingInterval(1*time.Second))

	config := &Configuration{
		Campaigns: []*Campaign{{
			ID: "test_cid",
		}},
	}

	engine.apiClient = NewAPIClientMock(testEnvID, config, 200)
	time.Sleep(1100 * time.Millisecond)

	assert.Equal(t, 1, len(engine.config.Campaigns))
	assert.Equal(t, false, engine.config.Panic)

	// Setting panic
	config.Panic = true

	time.Sleep(1100 * time.Millisecond)

	assert.Equal(t, 1, len(engine.config.Campaigns))
	assert.Equal(t, true, engine.config.Panic)
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	eg := utils.NewExecGroup(ctx)
	engine, _ := NewEngine(testEnvID, eg, PollingInterval(1*time.Second))

	config := &Configuration{
		Campaigns: []*Campaign{{
			ID: "test_cid",
		}},
	}

	engine.apiClient = NewAPIClientMock(testEnvID, config, 200)
	time.Sleep(1100 * time.Millisecond)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	eg.TerminateAndWait()
	wg.Wait()
}
