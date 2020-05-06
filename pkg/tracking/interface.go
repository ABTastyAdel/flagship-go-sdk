package tracking

// APIClientInterface sends a hit to the data collect
type APIClientInterface interface {
	sendInternalHit(hit HitInterface) error
	ActivateCampaign(request ActivationHit) error
}

// HitInterface express the interface for the hits
type HitInterface interface {
	validate() []error
	setBaseInfos(envID string, visitorID string)
	getBaseHit() BaseHit
	resetBaseHit()
	computeQueueTime()
}
