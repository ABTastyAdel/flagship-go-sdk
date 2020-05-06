package decision

import "time"

// APIOptions represents the options for the Decision API Client
type APIOptions struct {
	APIUrl  string
	APIKey  string
	Timeout time.Duration
	Retries int
}

// APIClientRequest represents the API client informations
type APIClientRequest struct {
	VisitorID  string                 `json:"visitor_id"`
	Context    map[string]interface{} `json:"context"`
	TriggerHit bool                   `json:"trigger_hit"`
}

// APIClientResponse represents a decision response
type APIClientResponse struct {
	VisitorID string              `json:"visitorId"`
	Panic     bool                `json:"panic"`
	Campaigns []APIClientCampaign `json:"campaigns"`
}

// APIClientCampaign represents a decision campaign
type APIClientCampaign struct {
	ID               string             `json:"id"`
	VariationGroupID string             `json:"variationGroupId"`
	Variation        APIClientVariation `json:"variation"`
}

// APIClientVariation represents a decision campaign variation
type APIClientVariation struct {
	ID            string                `json:"id"`
	Modifications APIClientModification `json:"modifications"`
	Reference     bool                  `json:"reference"`
}

// APIClientModification represents a decision campaign variation modification
type APIClientModification struct {
	Type  string                 `json:"type"`
	Value map[string]interface{} `json:"value"`
}

// APIClientFlagInfos represents a decision campaign variation modification
type APIClientFlagInfos struct {
	Value    interface{}
	Campaign APIClientCampaign
}
