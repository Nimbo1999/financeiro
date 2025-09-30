package proxy

// InMemoryServiceRegistry implements ServiceRegistry interface
type InMemoryServiceRegistry struct {
	services []ServiceConfig
}

// NewInMemoryServiceRegistry creates a new InMemoryServiceRegistry
func NewInMemoryServiceRegistry(services []ServiceConfig) *InMemoryServiceRegistry {
	return &InMemoryServiceRegistry{
		services: services,
	}
}

// GetServices returns all configured services
func (sr *InMemoryServiceRegistry) GetServices() []ServiceConfig {
	return sr.services
}

// AddService adds a new service to the registry
func (sr *InMemoryServiceRegistry) AddService(service ServiceConfig) {
	sr.services = append(sr.services, service)
}