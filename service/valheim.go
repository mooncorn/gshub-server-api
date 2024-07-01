package service

import (
	"errors"

	"github.com/mooncorn/gshub-server-api/cycles"
)

type ValheimServiceStrategy struct {
	data *cycles.ServiceData
}

func NewValheimServiceStrategy(data *cycles.ServiceData) *ValheimServiceStrategy {
	return &ValheimServiceStrategy{
		data: data,
	}
}

func (s *ValheimServiceStrategy) CreateBaseConfig() map[string]string {
	return make(map[string]string)
}

func (s *ValheimServiceStrategy) FormatCommand(cmd string) (string, error) {
	return "", errors.New("feature not supported")
}
