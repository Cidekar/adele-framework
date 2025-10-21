package provider

// Provider manages the loading and bootstrapping of service providers
type Provider struct {
	//App              *adele.Adele
	EnabledProviders map[string]bool
	ProviderConfigs  map[string]map[string]interface{}
}

// ServiceProvider is the expected interface every provider must implement
type ServiceProvider interface {
	Register(app interface{}) error
	Boot(app interface{}) error
	Name() string
}

// OptionalProvider allows providers to specify if they're optional
type OptionalProvider interface {
	ServiceProvider
	IsOptional() bool
}

// ConfigurableProvider allows providers to accept configuration
type ConfigurableProvider interface {
	ServiceProvider
	Configure(config map[string]interface{}) error
}

// PriorityProvider allows providers to specify boot order
type PriorityProvider interface {
	ServiceProvider
	Priority() int
}
