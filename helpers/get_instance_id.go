package helpers

import (
	"os"
)

var INSTANCE_ID string

func GetInstanceId() {
	var instanceID string

	if testID := os.Getenv("INSTANCE_ID"); testID != "" {
		instanceID = testID
	} else {
		realID, err := ExecuteCommand("echo", "$INSTANCE_ID")
		if err != nil {
			panic(err)
		}
		instanceID = realID
	}

	INSTANCE_ID = instanceID
}
