package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mooncorn/gshub-core/models"
)

func CalculateServerMemory(maxAvailableMemoryGB int) int {
	return maxAvailableMemoryGB - 1 // 1G allocated for the system and api
}

func CheckMinimumMemory(serverMemoryGB int, minimumServerMemoryGB int) bool {
	return minimumServerMemoryGB <= serverMemoryGB
}

// Relies on the name of env value which has to contain the number in GiB
func FindEnvMemoryValue(service models.Service, serverMemoryGB int, memoryName string) (models.ServiceEnvValue, error) {
	for _, env := range service.Env {
		if env.Name == memoryName {
			for _, envValue := range env.Values {
				if strings.Contains(envValue.Name, strconv.Itoa(serverMemoryGB)) {
					return envValue, nil
				}
			}
			break
		}
	}

	return models.ServiceEnvValue{}, fmt.Errorf("memory not found in env values (%d GiB)", serverMemoryGB)
}
