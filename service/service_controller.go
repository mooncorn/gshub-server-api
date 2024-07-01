package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	coreConfig "github.com/mooncorn/gshub-core/config"
	"github.com/mooncorn/gshub-server-api/internal"
)

type InstanceData struct {
	internal.StartupPayload
	InstanceID string
}

type ServiceController struct {
	docker         *DockerClient
	data           *InstanceData
	serviceFactory ServiceStrategyFactory
}

const SERVICE_CONTAINER_ID = "main"

func NewServiceController(data *InstanceData) (*ServiceController, error) {
	docker, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker container: %v", err)
	}

}

// check for existing container and return appropriate strategy for it
func (s *ServiceController) getStrategy(c context.Context) (*ServiceStrategy, error) {
	container, err := s.docker.GetContainer(c, "main")
	if err != nil {
		return nil, fmt.Errorf("failed to get main container: %v", err)
	}

	// find serviceNameID using image from container
	for _, conf := range s.data.ServiceConfigs {
		if strings.EqualFold(conf.Image, container.Image) {
			return s.serviceFactory.CreateService(conf.Name)
		}
	}

	return nil, fmt.Errorf("no strategy found for this container image: %s", container.Image)
}

func (s *ServiceController) CreateService(c context.Context, serviceNameID string, serviceEnv map[string]string) (*Container, error) {
	// check if there's already a service created
	if _, err := s.docker.GetContainer(c, SERVICE_CONTAINER_ID); err == nil {
		return nil, fmt.Errorf("conflict: a service already exists on this instance")
	}

}

func (v *ServiceController) ValidateConfig(config map[string]string) (map[string]string, error) {
	// Verify if the plan can accommodate this type of service
	if !v.hasEnoughMemory() {
		return nil, errors.New("not supported: this instance does not meet minimum memory requirements for this service")
	}

	baseConfig := v.createBaseConfig()

	for _, env := range v.config.Env {
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
func (v *ServiceController) createBaseConfig() map[string]string {
	return v.strategy.CreateBaseConfig()
}

func (v *ServiceController) FormatGameCommand(cmd string) (string, error) {
	return v.strategy.FormatCommand(cmd)
}

// Check if the provided value is one of the allowed values.
func isValidConfigValue(values []coreConfig.Value, value string) bool {
	for _, envValue := range values {
		if envValue.Value == value {
			return true
		}
	}
	return false
}

func CalculateServiceMemory(instanceMemoryMB int) int {
	return instanceMemoryMB - 1024 // 1GB allocated for the system and api
}

func (c *ServiceController) hasEnoughMemory() bool {
	return c.data.ServiceMinimumMemoryRequired <= CalculateServiceMemory(c.data.InstanceMemory)
}
