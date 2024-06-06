package server_strategies

import "github.com/mooncorn/gshub-core/models"

// Strategy defines the interface for different strategies
type ServerStrategy interface {
	CreateBaseConfig(service *models.Service, plan *models.Plan) map[string]string
	FormatCommand(cmd string) (string, error)
}

// StrategyFactory returns a map of available strategies
func GetServerStrategyFactory() map[string]ServerStrategy {
	return map[string]ServerStrategy{
		"minecraft": &MinecraftServerStrategy{},
		"valheim":   &ValheimServerStrategy{},
	}
}
