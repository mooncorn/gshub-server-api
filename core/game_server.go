package core

import (
	"errors"
	"fmt"

	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/core/memory"
	"github.com/mooncorn/gshub-server-api/core/server_strategies"
)

// Context holds a reference to a Strategy
type GameServer struct {
	service  *models.Service
	plan     *models.Plan
	strategy server_strategies.ServerStrategy
}

func NewGameServer(service *models.Service, plan *models.Plan) (*GameServer, error) {
	strategies := server_strategies.GetServerStrategyFactory()

	strategy, exists := strategies[service.NameID]
	if !exists {
		return nil, fmt.Errorf("no server strategy for this service: %s", service.NameID)
	}

	return &GameServer{
		service:  service,
		plan:     plan,
		strategy: strategy,
	}, nil
}

func (v *GameServer) ValidateConfig(config map[string]string) (map[string]string, error) {
	// Verify if the plan can accommodate this type of service
	serverMemory := memory.CalculateServerMemory(v.plan.Memory)
	if !memory.CheckMinimumMemory(serverMemory, v.service.MinMem) {
		return nil, errors.New("not enough memory for this service")
	}

	baseConfig := v.createBaseConfig()

	for _, env := range v.service.Env {
		value, ok := config[env.Key]

		if !ok {
			if env.Required {
				return nil, fmt.Errorf("%s is required", env.Key)
			}

			baseConfig[env.Key] = env.Default
			continue
		}

		if !isValidConfigValue(env.Values, value) {
			return nil, fmt.Errorf("invalid %s: %s", env.Key, value)
		}

		baseConfig[env.Key] = value
	}

	return baseConfig, nil
}

// ExecuteStrategy executes the current strategy
func (v *GameServer) createBaseConfig() map[string]string {
	return v.strategy.CreateBaseConfig(v.service, v.plan)
}

func (v *GameServer) FormatGameCommand(cmd string) (string, error) {
	return v.strategy.FormatCommand(cmd)
}

// Check if the provided value is one of the allowed values.
func isValidConfigValue(values []models.ServiceEnvValue, value string) bool {
	for _, envValue := range values {
		if envValue.Value == value {
			return true
		}
	}
	return false
}
