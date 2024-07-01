package service

import (
	"fmt"

	"github.com/mooncorn/gshub-server-api/cycles"
)

type MinecraftServiceStrategy struct {
	data *cycles.ServiceData
}

func NewMinecraftServiceStrategy(data *cycles.ServiceData) *MinecraftServiceStrategy {
	return &MinecraftServiceStrategy{
		data: data,
	}
}

func (s *MinecraftServiceStrategy) CreateBaseConfig() map[string]string {
	serviceMemory := CalculateServiceMemory(s.data.InstanceMemory)

	return map[string]string{
		"MEMORY": fmt.Sprintf("%dM", serviceMemory),
		"EULA":   "TRUE",
	}
}

func (s *MinecraftServiceStrategy) FormatCommand(cmd string) (string, error) {
	return fmt.Sprintf("rcon-cli %s", cmd), nil
}
