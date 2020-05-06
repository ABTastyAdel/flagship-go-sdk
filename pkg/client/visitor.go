package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/pkg/tracking"
	"github.com/abtasty/flagship-go-sdk/pkg/utils"
)

var visitorLogger = logging.GetLogger("FS Visitor")

// FlagshipVisitor is the entry point to the Flagship SDK
type FlagshipVisitor struct {
	ID                string
	Context           map[string]interface{}
	decisionClient    decision.ClientInterface
	decisionResponse  *decision.APIClientResponse
	flagInfos         map[string]decision.APIClientFlagInfos
	trackingAPIClient tracking.APIClientInterface
	batchHitProcessor *tracking.BatchHitProcessor
}

// UpdateContext updates the FlagshipVisitor context with new value
func (v *FlagshipVisitor) UpdateContext(newContext map[string]interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	errs := validateContext(newContext)
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			visitorLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	v.Context = newContext
	return nil
}

// UpdateContextKey updates a single FlagshipVisitor context key with new value
func (v *FlagshipVisitor) UpdateContextKey(key string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	newContext := map[string]interface{}{}
	for k, v := range v.Context {
		newContext[k] = v
	}

	newContext[key] = value

	errs := validateContext(newContext)
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			visitorLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	v.Context = newContext
	return nil
}

// SynchronizeModifications updates the latest campaigns and modifications for the visitor
func (v *FlagshipVisitor) SynchronizeModifications() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	if v.ID == "" {
		err := errors.New("Visitor ID should not be empty")
		visitorLogger.Error("Visitor ID is not set", err)
		return err
	}

	visitorLogger.Info(fmt.Sprintf("Getting modifications for visitor with id : %s", v.ID))
	resp, err := v.decisionClient.GetModifications(v.ID, v.Context)

	if err != nil {
		visitorLogger.Error("Error when calling Decision API", errors.New("Visitor ID should not be empty"))
		return err
	}
	v.decisionResponse = resp

	v.flagInfos = map[string]decision.APIClientFlagInfos{}

	visitorLogger.Info(fmt.Sprintf("Got %d campaign(s) for visitor with id : %s", len(resp.Campaigns), v.ID))
	for _, c := range resp.Campaigns {
		for k, val := range c.Variation.Modifications.Value {
			v.flagInfos[k] = decision.APIClientFlagInfos{
				Value:    val,
				Campaign: c,
			}
		}
	}

	return nil
}

// getModification gets a flag value as interface{}
func (v *FlagshipVisitor) getModification(key string, activate bool) (flagValue interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	if v.flagInfos == nil {
		err := errors.New("Visitor modifications have not been synchronized")
		visitorLogger.Error("Visitor modifications are not set", err)

		return false, err
	}

	flagInfos, ok := v.flagInfos[key]

	if !ok {
		return nil, fmt.Errorf("Key %s not set in decision infos. Fallback to default value", key)
	}

	if activate {
		visitorLogger.Info(fmt.Sprintf("Activating campaign for flag %s for visitor with id : %s", key, v.ID))
		err := v.trackingAPIClient.ActivateCampaign(tracking.ActivationHit{
			VariationGroupID: flagInfos.Campaign.VariationGroupID,
			VariationID:      flagInfos.Campaign.Variation.ID,
			VisitorID:        v.ID,
		})

		if err != nil {
			visitorLogger.Debug(fmt.Sprintf("Error occurred when activating campaign : %v.", err))
		}
		// ok := v.batchHitProcessor.ProcessHit(v.ID, &tracking.ActivationHit{
		// 	VariationGroupID: flagInfos.Campaign.VariationGroupID,
		// 	VariationID:      flagInfos.Campaign.Variation.ID,
		// })

		// if !ok {
		// 	visitorLogger.Debug(fmt.Sprintf("Error occurred when activating campaign : %v.", err))
		// 	err = errors.New("Error when registering hit")
		// }
	}
	flagValue = flagInfos.Value
	return flagValue, nil
}

// GetAllModifications return all the modifications
func (v *FlagshipVisitor) GetAllModifications() (flagInfos map[string]decision.APIClientFlagInfos) {
	return v.flagInfos
}

// GetModificationBool get a modification bool by its key
func (v *FlagshipVisitor) GetModificationBool(key string, defaultValue bool, activate bool) (castVal bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(bool)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type bool. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected bool, got %v", val)
	}

	return castVal, nil
}

// GetModificationString get a modification string by its key
func (v *FlagshipVisitor) GetModificationString(key string, defaultValue string, activate bool) (castVal string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(string)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type string. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected string, got %v", val)
	}

	return castVal, nil
}

// GetModificationNumber get a modification number as float64 by its key
func (v *FlagshipVisitor) GetModificationNumber(key string, defaultValue float64, activate bool) (castVal float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(float64)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type float. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected float64, got %v", val)
	}

	return castVal, nil
}

// ActivateModification notifies Flagship that the visitor has seen to modification
func (v *FlagshipVisitor) ActivateModification(key string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	_, err = v.getModification(key, true)

	return err
}

// SendHit sends a tracking hit to the Data Collect API
func (v *FlagshipVisitor) SendHit(hit tracking.HitInterface) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	visitorLogger.Info(fmt.Sprintf("Sending hit for visitor with id : %s", v.ID))
	ok := v.batchHitProcessor.ProcessHit(v.ID, hit)

	if !ok {
		err = errors.New("Error when registering hit")
	}
	return err
}
