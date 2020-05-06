package decision

// ClientInterface sends a hit to the data collect
type ClientInterface interface {
	GetModifications(visitorID string, context map[string]interface{}) (*APIClientResponse, error)
}
