package helpers

import (
	"bytes"
	"os/exec"
)

// Executes the command inside the server-api container on an instance
func ExecuteCommand(name string, arg ...string) (string, error) {
	// Create a new command
	cmd := exec.Command(name, arg...)

	// Create a buffer to store the command output
	var out bytes.Buffer

	// Set the output stream of the command to our buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Execute the command
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Convert the buffer to a string and return
	return out.String(), nil
}
