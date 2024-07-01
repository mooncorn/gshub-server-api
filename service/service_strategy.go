package service

import (
	"fmt"

	"github.com/mooncorn/gshub-server-api/internal"
)

// Defines the interface for different strategies
type ServiceStrategy interface {
	CreateBaseConfig() map[string]string
	FormatCommand(cmd string) (string, error)
}

type ServiceStrategyFactory interface {
	CreateService(serviceNameID string) (*ServiceStrategy, error)
}

type ServiceFactory struct {
	data *map[string]internal.ServiceConfiguration
}

func NewServiceFactory(data *map[string]internal.ServiceConfiguration) *ServiceFactory {
	return &ServiceFactory{
		data: data,
	}
}

// Returns a map of available strategies
func (s *ServiceFactory) CreateService(serviceNameID string) (*ServiceStrategy, error) {
	strats := map[string]ServiceStrategy{
		"minecraft": NewMinecraftServiceStrategy(s.data),
		"valheim":   NewValheimServiceStrategy(s.data),
	}

	strat, exists := strats[serviceNameID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceNameID)
	}

	return &strat, nil
}
