package client

import (
	"fmt"

	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

var logger = logging.GetLogger("FS Factory")

// DecisionMode express a targeting operator
type DecisionMode string

// The different targeting operators
const (
	API       DecisionMode = "API"
	Bucketing DecisionMode = "Bucketing"
)

// FlagshipFactory is the entry point to the Flagship SDK
type FlagshipFactory struct {
	EnvID              string
	decisionMode       DecisionMode
	bucketingOptions   []func(*bucketing.Engine)
	decisionAPIOptions []func(*decision.APIClient)
}

// OptionFunc is a func type to set options to the FlagshipFactory.
type OptionFunc func(*FlagshipFactory)

// CreateClient creates a FlagshipClient from envID and options
func (f *FlagshipFactory) CreateClient(clientOptions ...OptionFunc) (*FlagshipClient, error) {
	f.decisionMode = API

	// extract options
	for _, opt := range clientOptions {
		opt(f)
	}

	logger.Info(fmt.Sprintf("Creating FS Client with Decision Mode : %s", f.decisionMode))
	client := &FlagshipClient{
		envID:              f.EnvID,
		decisionMode:       f.decisionMode,
		bucketingOptions:   f.bucketingOptions,
		decisionAPIOptions: f.decisionAPIOptions,
	}
	client.init()

	return client, nil
}

// WithBucketing enables the bucketing decision mode for the SDK
func WithBucketing(options ...func(*bucketing.Engine)) OptionFunc {
	return func(f *FlagshipFactory) {
		f.decisionMode = Bucketing
		f.bucketingOptions = options
	}
}

// WithDecisionAPI changes the decision API options
func WithDecisionAPI(options ...func(*decision.APIClient)) OptionFunc {
	return func(f *FlagshipFactory) {
		f.decisionAPIOptions = options
	}
}
