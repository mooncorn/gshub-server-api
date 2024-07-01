package system

import "os/exec"

type UnixCommandRunner struct{}

func NewUnixCommandRunner() *UnixCommandRunner {
	return &UnixCommandRunner{}
}

func (cr *UnixCommandRunner) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (cr *UnixCommandRunner) RunAsAdmin(command string, args ...string) (string, error) {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("sudo", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
