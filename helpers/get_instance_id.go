package helpers

import (
	"fmt"
	"os"
)

func GetInstanceId() (string, error) {
	var instanceID string

	if testID := os.Getenv("INSTANCE_ID"); testID != "" {
		instanceID = testID
	} else {
		realID, err := ExecuteCommand("echo", "$INSTANCE_ID")
		if err != nil {
			return "", fmt.Errorf("failed to identify instance: %v", err)
		}
		instanceID = realID
	}

	return instanceID, nil
}
