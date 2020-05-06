package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/abtasty/flagship-go-sdk/pkg/utils"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/pkg/tracking"
)

// FlagshipClient is the entry point to the Flagship SDK
type FlagshipClient struct {
	envID              string
	decisionMode       DecisionMode
	decisionClient     decision.ClientInterface
	decisionAPIOptions []func(*decision.APIClient)
	trackingAPIClient  tracking.APIClientInterface
	bucketingOptions   []func(*bucketing.Engine)
	batchHitProcessor  *tracking.BatchHitProcessor
	executionGroup     *utils.ExecGroup
}

var clientLogger = logging.GetLogger("FS Client")

// init the tracking and decision clients
func (c *FlagshipClient) init() {
	eg := utils.NewExecGroup(context.Background())
	c.executionGroup = eg

	if c.decisionClient == nil {
		if c.decisionMode == Bucketing {
			var err error
			c.decisionClient, err = bucketing.NewEngine(c.envID, c.executionGroup, c.bucketingOptions...)
			if err != nil {
				clientLogger.Error("Got error when creating bucketing engine", err)
			}
		} else {
			c.decisionClient = decision.NewAPIClient(c.envID, c.decisionAPIOptions...)
		}
	}
	if c.trackingAPIClient == nil {
		c.trackingAPIClient = tracking.NewAPIClient(c.envID)
	}

	c.batchHitProcessor = tracking.NewBatchHitProcessor(c.envID)
	eg.Go(c.batchHitProcessor.Start)

}

func validateContext(context map[string]interface{}) []error {
	errorList := []error{}
	for key, val := range context {
		_, okBool := val.(bool)
		_, okString := val.(string)
		_, okFloat64 := val.(float64)
		intVal, okInt := val.(int)

		if !okBool && !okString && !okFloat64 && !okInt {
			errorList = append(errorList, fmt.Errorf("Value %v not handled for key %s. Type must be one of string, bool or number (int or float64)", val, key))
		}

		if okInt {
			context[key] = float64(intVal)
		}
	}
	return errorList
}

// NewVisitor returns a new FlagshipVisitor from ID and context
func (c *FlagshipClient) NewVisitor(visitorID string, context map[string]interface{}) (visitor *FlagshipVisitor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, clientLogger)
		}
	}()

	clientLogger.Info(fmt.Sprintf("Creating new visitor with id : %s", visitorID))

	errs := validateContext(context)
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			clientLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return nil, fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	return &FlagshipVisitor{
		ID:                visitorID,
		Context:           context,
		decisionClient:    c.decisionClient,
		batchHitProcessor: c.batchHitProcessor,
		trackingAPIClient: c.trackingAPIClient,
	}, nil
}

// SendHit sends a tracking hit to the Data Collect API
func (c *FlagshipClient) SendHit(visitorID string, hit tracking.HitInterface) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, clientLogger)
		}
	}()

	clientLogger.Info(fmt.Sprintf("Sending hit for visitor with id : %s", visitorID))
	ok := c.batchHitProcessor.ProcessHit(visitorID, hit)

	if !ok {
		err = errors.New("Error when registering hit")
	}
	return err
}

// Dispose disposes the FlagshipClient and close all connections
func (c *FlagshipClient) Dispose() (err error) {
	c.executionGroup.TerminateAndWait()
	return err
}
