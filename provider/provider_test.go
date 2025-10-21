package provider

import (
	"errors"
	"testing"
)

// Mock providers for testing
type mockProvider struct {
	name         string
	priority     int
	isOptional   bool
	registerErr  error
	bootErr      error
	configured   bool
	registered   bool
	booted       bool
	configValues map[string]interface{}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Register(app interface{}) error {
	m.registered = true
	return m.registerErr
}

func (m *mockProvider) Boot(app interface{}) error {
	m.booted = true
	return m.bootErr
}

func (m *mockProvider) Priority() int {
	return m.priority
}

func (m *mockProvider) IsOptional() bool {
	return m.isOptional
}

func (m *mockProvider) Configure(config map[string]interface{}) error {
	m.configured = true
	m.configValues = config
	return nil
}

// Reset global providers before each test
func resetGlobalProviders() {
	globalProviders = []ServiceProvider{}
}

func TestRegisterGlobalProvider(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "test1"}
	provider2 := &mockProvider{name: "test2"}

	RegisterGlobalProvider(provider1)
	RegisterGlobalProvider(provider2)

	providers := GetRegisteredProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
}

func TestRegisterGlobalProviderDuplicate(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "duplicate"}
	provider2 := &mockProvider{name: "duplicate"}

	RegisterGlobalProvider(provider1)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate provider name")
		}
	}()

	RegisterGlobalProvider(provider2)
}

func TestGetRegisteredProviders(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "test1"}
	RegisterGlobalProvider(provider1)

	providers := GetRegisteredProviders()

	// Modify the returned slice
	providers = append(providers, &mockProvider{name: "test2"})

	// Original should be unchanged
	original := GetRegisteredProviders()
	if len(original) != 1 {
		t.Error("GetRegisteredProviders should return a copy")
	}
}

func TestProviderSetProviderEnabled(t *testing.T) {
	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	p.SetProviderEnabled("test", false)

	if p.EnabledProviders["test"] != false {
		t.Error("Provider should be disabled")
	}
}

func TestProviderIsProviderEnabled(t *testing.T) {
	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	// Default should be enabled
	if !p.IsProviderEnabled("unknown") {
		t.Error("Unknown provider should be enabled by default")
	}

	// Explicitly disabled
	p.SetProviderEnabled("disabled", false)
	if p.IsProviderEnabled("disabled") {
		t.Error("Disabled provider should return false")
	}

	// Explicitly enabled
	p.SetProviderEnabled("enabled", true)
	if !p.IsProviderEnabled("enabled") {
		t.Error("Enabled provider should return true")
	}
}

func TestProviderSetProviderConfig(t *testing.T) {
	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	config := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	p.SetProviderConfig("test", config)

	retrieved := p.GetProviderConfig("test")
	if retrieved["key1"] != "value1" {
		t.Error("Config value mismatch")
	}
	if retrieved["key2"] != 42 {
		t.Error("Config value mismatch")
	}
}

func TestProviderGetProviderConfig(t *testing.T) {
	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	// Non-existent config should return nil
	if config := p.GetProviderConfig("nonexistent"); config != nil {
		t.Error("Non-existent config should return nil")
	}
}

func TestLoadProvidersBasic(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "test1"}
	provider2 := &mockProvider{name: "test2"}

	RegisterGlobalProvider(provider1)
	RegisterGlobalProvider(provider2)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err != nil {
		t.Errorf("LoadProviders failed: %v", err)
	}

	if !provider1.registered || !provider1.booted {
		t.Error("Provider1 should be registered and booted")
	}
	if !provider2.registered || !provider2.booted {
		t.Error("Provider2 should be registered and booted")
	}
}

func TestLoadProvidersDisabled(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "enabled"}
	provider2 := &mockProvider{name: "disabled"}

	RegisterGlobalProvider(provider1)
	RegisterGlobalProvider(provider2)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	p.SetProviderEnabled("disabled", false)

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err != nil {
		t.Errorf("LoadProviders failed: %v", err)
	}

	if !provider1.registered {
		t.Error("Enabled provider should be registered")
	}
	if provider2.registered {
		t.Error("Disabled provider should not be registered")
	}
}

func TestLoadProvidersWithConfiguration(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "configurable"}

	RegisterGlobalProvider(provider1)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	config := map[string]interface{}{
		"key": "value",
	}
	p.SetProviderConfig("configurable", config)

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err != nil {
		t.Errorf("LoadProviders failed: %v", err)
	}

	if !provider1.configured {
		t.Error("Provider should be configured")
	}
	if provider1.configValues["key"] != "value" {
		t.Error("Config values not passed correctly")
	}
}

func TestLoadProvidersRegisterError(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{
		name:        "failing",
		registerErr: errors.New("register failed"),
	}

	RegisterGlobalProvider(provider1)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err == nil {
		t.Error("Expected error from failing provider")
	}
}

func TestLoadProvidersBootError(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{
		name:    "failing",
		bootErr: errors.New("boot failed"),
	}

	RegisterGlobalProvider(provider1)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err == nil {
		t.Error("Expected error from failing provider boot")
	}
}

func TestLoadProvidersOptionalBootError(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{
		name:       "optional",
		bootErr:    errors.New("boot failed"),
		isOptional: true,
	}

	RegisterGlobalProvider(provider1)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err != nil {
		t.Error("Optional provider failure should not stop loading")
	}
}

func TestSortProvidersByPriorityDefault(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "default1"}
	provider2 := &mockProvider{name: "default2"}
	provider3 := &mockProvider{name: "default3"}

	providers := []ServiceProvider{provider1, provider2, provider3}

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	sorted := p.sortProvidersByPriority(providers)

	// All have default priority, order should be preserved
	if len(sorted) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(sorted))
	}
}

func TestSortProvidersByPriorityWithPriorityInterface(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "high", priority: 10}
	provider2 := &mockProvider{name: "low", priority: 90}
	provider3 := &mockProvider{name: "medium", priority: 50}

	providers := []ServiceProvider{provider2, provider1, provider3}

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	sorted := p.sortProvidersByPriority(providers)

	// Should be sorted: high (10), medium (50), low (90)
	if sorted[0].Name() != "high" {
		t.Errorf("Expected 'high' first, got %s", sorted[0].Name())
	}
	if sorted[1].Name() != "medium" {
		t.Errorf("Expected 'medium' second, got %s", sorted[1].Name())
	}
	if sorted[2].Name() != "low" {
		t.Errorf("Expected 'low' third, got %s", sorted[2].Name())
	}
}

func TestSortProvidersByPriorityWithConfig(t *testing.T) {
	resetGlobalProviders()

	provider1 := &mockProvider{name: "provider1", priority: 50}
	provider2 := &mockProvider{name: "provider2", priority: 60}

	providers := []ServiceProvider{provider1, provider2}

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	// Override provider1's priority via config
	p.SetProviderConfig("provider1", map[string]interface{}{
		"priority": 80,
	})

	sorted := p.sortProvidersByPriority(providers)

	// provider2 (60) should come before provider1 (80 from config)
	if sorted[0].Name() != "provider2" {
		t.Errorf("Expected 'provider2' first, got %s", sorted[0].Name())
	}
	if sorted[1].Name() != "provider1" {
		t.Errorf("Expected 'provider1' second, got %s", sorted[1].Name())
	}
}

// noPriorityProvider is a provider without Priority() method
type noPriorityProvider struct {
	name       string
	registered bool
	booted     bool
}

func (n *noPriorityProvider) Name() string {
	return n.name
}

func (n *noPriorityProvider) Register(app interface{}) error {
	n.registered = true
	return nil
}

func (n *noPriorityProvider) Boot(app interface{}) error {
	n.booted = true
	return nil
}

func TestSortProvidersByPriorityMixed(t *testing.T) {
	resetGlobalProviders()

	// Provider with Priority() method
	provider1 := &mockProvider{name: "hasPriority", priority: 30}
	// Provider without Priority() method (default 100)
	provider2 := &noPriorityProvider{name: "default"}
	// Provider with config override
	provider3 := &mockProvider{name: "configOverride", priority: 70}

	providers := []ServiceProvider{provider2, provider3, provider1}

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	// Override provider3's priority
	p.SetProviderConfig("configOverride", map[string]interface{}{
		"priority": 10,
	})

	sorted := p.sortProvidersByPriority(providers)

	// Should be: configOverride (10), hasPriority (30), default (100)
	if sorted[0].Name() != "configOverride" {
		t.Errorf("Expected 'configOverride' first, got %s", sorted[0].Name())
	}
	if sorted[1].Name() != "hasPriority" {
		t.Errorf("Expected 'hasPriority' second, got %s", sorted[1].Name())
	}
	if sorted[2].Name() != "default" {
		t.Errorf("Expected 'default' third, got %s", sorted[2].Name())
	}
}

// trackingProvider is a provider that tracks execution order
type trackingProvider struct {
	name           string
	priority       int
	executionOrder *[]string
	orderPrefix    string
}

func (t *trackingProvider) Name() string {
	return t.name
}

func (t *trackingProvider) Priority() int {
	return t.priority
}

func (t *trackingProvider) Register(app interface{}) error {
	*t.executionOrder = append(*t.executionOrder, t.orderPrefix+"-register")
	return nil
}

func (t *trackingProvider) Boot(app interface{}) error {
	*t.executionOrder = append(*t.executionOrder, t.orderPrefix+"-boot")
	return nil
}

func TestLoadProvidersExecutionOrder(t *testing.T) {
	resetGlobalProviders()

	var executionOrder []string

	provider1 := &trackingProvider{
		name:           "first",
		priority:       10,
		executionOrder: &executionOrder,
		orderPrefix:    "first",
	}
	provider2 := &trackingProvider{
		name:           "second",
		priority:       50,
		executionOrder: &executionOrder,
		orderPrefix:    "second",
	}
	provider3 := &trackingProvider{
		name:           "third",
		priority:       90,
		executionOrder: &executionOrder,
		orderPrefix:    "third",
	}

	// Register in non-priority order
	RegisterGlobalProvider(provider3)
	RegisterGlobalProvider(provider1)
	RegisterGlobalProvider(provider2)

	p := &Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	app := &struct{}{}
	err := p.LoadProviders(app)
	if err != nil {
		t.Errorf("LoadProviders failed: %v", err)
	}

	// Check that they registered in priority order, then booted in same order
	expected := []string{
		"first-register",
		"second-register",
		"third-register",
		"first-boot",
		"second-boot",
		"third-boot",
	}

	if len(executionOrder) != len(expected) {
		t.Errorf("Expected %d executions, got %d", len(expected), len(executionOrder))
	}

	for i, expectedName := range expected {
		if i >= len(executionOrder) {
			break
		}
		if executionOrder[i] != expectedName {
			t.Errorf("Execution order mismatch at position %d: expected %s, got %s", i, expectedName, executionOrder[i])
		}
	}
}
