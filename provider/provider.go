package provider

import (
	"fmt"
)

var globalProviders []ServiceProvider

// RegisterGlobalProvider adds a provider to the global registry
func RegisterGlobalProvider(provider ServiceProvider) {
	for _, p := range globalProviders {
		if p.Name() == provider.Name() {
			panic(fmt.Sprintf("provider with name '%s' already registered", provider.Name()))
		}
	}
	globalProviders = append(globalProviders, provider)
}

// GetRegisteredProviders returns a copy of all registered providers
func GetRegisteredProviders() []ServiceProvider {
	providers := make([]ServiceProvider, len(globalProviders))
	copy(providers, globalProviders)
	return providers
}

// IsProviderEnabled checks if a provider is enabled in configuration
func (p *Provider) IsProviderEnabled(name string) bool {
	enabled, exists := p.EnabledProviders[name]
	if !exists {
		return true // Default to enabled if not specified
	}
	return enabled
}

// GetProviderConfig returns configuration for a specific provider
func (p *Provider) GetProviderConfig(name string) map[string]interface{} {
	return p.ProviderConfigs[name]
}

// SetProviderEnabled enables or disables a provider
func (p *Provider) SetProviderEnabled(name string, enabled bool) {
	p.EnabledProviders[name] = enabled
}

// SetProviderConfig sets configuration for a provider
func (p *Provider) SetProviderConfig(name string, config map[string]interface{}) {
	fmt.Println("SetProviderConfig:", name, config)
	p.ProviderConfigs[name] = config
}

// LoadProviders discovers and loads all registered providers into the application
func (p *Provider) LoadProviders(app interface{}) error {
	providers := make([]ServiceProvider, len(globalProviders))
	copy(providers, globalProviders)

	// Sort providers by priority
	sortedProviders := p.sortProvidersByPriority(providers)

	// First pass: Register all enabled providers
	var registeredProviders []ServiceProvider
	for _, prov := range sortedProviders {
		if !p.IsProviderEnabled(prov.Name()) {
			fmt.Printf("Skipping disabled provider: %s\n", prov.Name())
			continue
		}

		// Configure provider if it supports configuration
		if configurable, ok := prov.(ConfigurableProvider); ok {
			if config := p.GetProviderConfig(prov.Name()); config != nil {
				if err := configurable.Configure(config); err != nil {
					return fmt.Errorf("failed to configure provider '%s': %w", prov.Name(), err)
				}
			}
		}

		fmt.Printf("Registering provider: %s\n", prov.Name())
		if err := prov.Register(app); err != nil {
			return fmt.Errorf("failed to register provider '%s': %w", prov.Name(), err)
		}

		registeredProviders = append(registeredProviders, prov)
	}

	// Second pass: Boot all registered providers
	for _, prov := range registeredProviders {
		fmt.Printf("Booting provider: %s\n", prov.Name())
		if err := prov.Boot(app); err != nil {
			// Check if provider is optional
			if optional, ok := prov.(OptionalProvider); ok && optional.IsOptional() {
				fmt.Printf("Warning: Optional provider '%s' failed to boot: %v\n", prov.Name(), err)
				continue
			}
			return fmt.Errorf("failed to boot provider '%s': %w", prov.Name(), err)
		}
	}

	fmt.Printf("Successfully loaded %d providers\n", len(registeredProviders))
	return nil
}

// sortProvidersByPriority sorts providers by priority (lowest first)
func (p *Provider) sortProvidersByPriority(providers []ServiceProvider) []ServiceProvider {
	type providerWithPriority struct {
		provider ServiceProvider
		priority int
	}

	withPriority := make([]providerWithPriority, len(providers))
	for i, prov := range providers {
		priority := 100 // default priority

		// First check if developer set priority in config
		if config := p.GetProviderConfig(prov.Name()); config != nil {
			if customPriority, ok := config["priority"].(int); ok {
				priority = customPriority
			} else if pp, ok := prov.(PriorityProvider); ok {
				// Fall back to provider's default priority
				priority = pp.Priority()
			}
		} else if pp, ok := prov.(PriorityProvider); ok {
			// No config, use provider's default priority
			priority = pp.Priority()
		}
		// Otherwise use default priority of 100

		withPriority[i] = providerWithPriority{prov, priority}
	}

	// Simple insertion sort
	for i := 1; i < len(withPriority); i++ {
		key := withPriority[i]
		j := i - 1
		for j >= 0 && withPriority[j].priority > key.priority {
			withPriority[j+1] = withPriority[j]
			j--
		}
		withPriority[j+1] = key
	}

	sorted := make([]ServiceProvider, len(providers))
	for i, prov := range withPriority {
		sorted[i] = prov.provider
	}
	return sorted
}
