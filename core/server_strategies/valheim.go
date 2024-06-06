package server_strategies

import (
	"errors"

	"github.com/mooncorn/gshub-core/models"
)

type ValheimServerStrategy struct{}

func (s *ValheimServerStrategy) CreateBaseConfig(service *models.Service, plan *models.Plan) map[string]string {
	return make(map[string]string)
}

func (s *ValheimServerStrategy) FormatCommand(cmd string) (string, error) {
	return "", errors.New("feature not supported")
}
