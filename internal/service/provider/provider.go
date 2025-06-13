package provider

import (
	"fmt"
	"sync"
)

type ServiceManager struct {
	providers map[string]ServiceI
	mu        sync.RWMutex
}

func NewManager() *ServiceManager {
	return &ServiceManager{
		providers: make(map[string]ServiceI),
	}
}

func (m *ServiceManager) RegisterProvider(providerID string, provider ServiceI) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if providerID == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}

	if _, exists := m.providers[providerID]; exists {
		return fmt.Errorf("provider with ID %s already registered", providerID)
	}

	m.providers[providerID] = provider

	return nil
}

func (m *ServiceManager) GetProvider(providerID string) (ServiceI, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider with ID %s not found", providerID)
	}

	return provider, nil
}

func (m *ServiceManager) ListProviders() []ServiceI {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]ServiceI, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}

	return providers
}
