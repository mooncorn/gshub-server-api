package system

import "fmt"

type CommandRunner interface {
	Run(command string, args ...string) (string, error)
	RunAsAdmin(command string, args ...string) (string, error)
}

type AmazonLinuxSystemController struct {
	commandRunner *UnixCommandRunner
}

func NewAmazonLinuxSystemController() *AmazonLinuxSystemController {
	return &AmazonLinuxSystemController{
		commandRunner: &UnixCommandRunner{},
	}
}

func (s *AmazonLinuxSystemController) GetInstanceID() (string, error) {
	command := "ec2-metadata"
	args := "-i | cut -d ' ' -f 2"
	output, err := s.commandRunner.Run(command, args)
	if err != nil {
		return "", fmt.Errorf("failed to get instance id: %v", err)
	}
	return output, nil
}

func (s *AmazonLinuxSystemController) Shutdown() (string, error) {
	command := "shutdown"
	output, err := s.commandRunner.RunAsAdmin(command)
	if err != nil {
		return "", fmt.Errorf("failed to get shutdown: %v", err)
	}
	return output, nil
}
