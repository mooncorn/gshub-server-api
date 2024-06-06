package server_strategies

import (
	"fmt"

	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/core/memory"
)

type MinecraftServerStrategy struct{}

func (s *MinecraftServerStrategy) CreateBaseConfig(service *models.Service, plan *models.Plan) map[string]string {
	serverMemory := memory.CalculateServerMemory(plan.Memory)

	return map[string]string{
		"MEMORY": fmt.Sprintf("%dM", serverMemory),
		"EULA":   "TRUE",
	}
}

func (s *MinecraftServerStrategy) FormatCommand(cmd string) (string, error) {
	return fmt.Sprintf("rcon-cli %s", cmd), nil
}
