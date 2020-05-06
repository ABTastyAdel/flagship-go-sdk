package bucketing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/utils"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

var logger = logging.GetLogger("Bucketing Engine")

// Engine represents a bucketing engine
type Engine struct {
	pollingInterval  time.Duration
	config           *Configuration
	apiClient        ConfigAPIInterface
	apiClientOptions []func(*APIClient)
	envID            string
	configMux        sync.Mutex
	executionGroup   *utils.ExecGroup
	ticker           *time.Ticker
}

// PollingInterval sets the polling interval for the bucketing engine
func PollingInterval(interval time.Duration) func(r *Engine) {
	return func(r *Engine) {
		r.pollingInterval = interval
	}
}

// APIOptions sets the func option for the engine client API Key
func APIOptions(apiOptions ...func(*APIClient)) func(r *Engine) {
	return func(r *Engine) {
		r.apiClientOptions = apiOptions
	}
}

// NewEngine creates a new engine for bucketing
func NewEngine(envID string, eg *utils.ExecGroup, params ...func(*Engine)) (*Engine, error) {
	engine := &Engine{
		pollingInterval:  1 * time.Minute,
		envID:            envID,
		apiClientOptions: []func(*APIClient){},
		executionGroup:   eg,
	}

	for _, param := range params {
		param(engine)
	}

	engine.apiClient = NewAPIClient(envID, engine.apiClientOptions...)

	err := engine.Load()

	if engine.pollingInterval != -1 {
		engine.executionGroup.Go(engine.startTicker)
	}

	return engine, err
}

// startTicker starts new ticker for flushing hits
func (b *Engine) startTicker(ctx context.Context) {
	if b.ticker != nil {
		return
	}
	b.ticker = time.NewTicker(b.pollingInterval)

	for {
		select {
		case <-b.ticker.C:
			logger.Info("Bucketing engine ticked, loading configuration")
			b.Load()
		case <-ctx.Done():
			logger.Info("Bucketing engine stopped")
			return
		}
	}
}

// Load loads the env configuration in cache
func (b *Engine) Load() error {
	newConfig, err := b.apiClient.GetConfiguration()

	if err != nil {
		logger.Error("Error when loading environment configuration", err)
		return err
	}

	b.configMux.Lock()
	b.config = newConfig
	b.configMux.Unlock()

	return nil
}

// GetModifications gets modifications from Decision API
func (b *Engine) GetModifications(visitorID string, context map[string]interface{}) (*decision.APIClientResponse, error) {
	if b.config == nil {
		logger.Info("Configuration not loaded. Loading it now")
		err := b.Load()
		if err != nil {
			logger.Warning("Configuration could not be loaded.")
			return nil, err
		}
	}

	resp := &decision.APIClientResponse{
		VisitorID: visitorID,
		Campaigns: []decision.APIClientCampaign{},
	}

	if b.config.Panic {
		logger.Info("Environment is in panic mode. Skipping all campaigns")
		return resp, nil
	}

	for _, c := range b.config.Campaigns {
		var matchedVg *VariationGroup
		for _, vg := range c.VariationGroups {
			matched, err := TargetingMatch(vg, visitorID, context)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error occured when checking targeting : %v", err))
				continue
			}

			if matched {
				matchedVg = vg
				break
			}
		}

		if matchedVg != nil {
			variation, err := GetRandomAllocation(visitorID, matchedVg)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error occured when allocating variation : %v", err))
				continue
			}
			campaign := decision.APIClientCampaign{
				ID:               c.ID,
				VariationGroupID: matchedVg.ID,
				Variation: decision.APIClientVariation{
					ID:        variation.ID,
					Reference: variation.Reference,
					Modifications: decision.APIClientModification{
						Type:  variation.Modifications.Type,
						Value: variation.Modifications.Value,
					},
				},
			}
			resp.Campaigns = append(resp.Campaigns, campaign)
		}
	}
	return resp, nil
}
