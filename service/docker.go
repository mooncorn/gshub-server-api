package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type PortBinding struct {
	Container string
	Host      string
	Protocol  string
}

type VolumeBinding struct {
	Container string
	Host      string
}

type Container struct {
	ID      string
	Image   string
	Running bool
	Name    string
	Env     map[string]string
	Ports   []PortBinding
	Volumes []VolumeBinding
}

type DockerClient struct {
	docker *client.Client
}

func NewDockerClient() (*DockerClient, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	return &DockerClient{
		docker: docker,
	}, nil
}

func (d *DockerClient) GetContainer(c context.Context, ID string) (Container, error) {
	container, err := d.docker.ContainerInspect(c, ID)
	if err != nil {
		return Container{}, fmt.Errorf("failed to inspect container: %v", err)
	}

	return mapToContainer(container), nil
}

func (d *DockerClient) CreateContainer() {

}

func mapToContainer(containerJSON types.ContainerJSON) Container {
	return Container{
		ID:      containerJSON.ID,
		Image:   containerJSON.Image,
		Running: containerJSON.State.Running,
		Name:    containerJSON.Name,
		Env:     mapToEnv(containerJSON.Config.Env),
		Volumes: mapToVolumes(containerJSON.HostConfig.Binds),
		Ports:   mapToPorts(containerJSON.HostConfig.PortBindings),
	}
}

func mapToEnv(envList []string) map[string]string {
	obj := make(map[string]string, len(envList))
	for _, envKeyValue := range envList {
		if envKeyValue == "" {
			continue
		}

		keyValue := strings.Split(envKeyValue, "=")

		if len(keyValue) != 2 {
			continue
		}

		key := keyValue[0]
		value := keyValue[1]

		obj[key] = value
	}
	return obj
}

func mapToVolumes(volumesArr []string) []VolumeBinding {
	var volumes []VolumeBinding
	for _, volumeStr := range volumesArr {
		if volumeStr == "" {
			continue
		}

		keyValue := strings.Split(volumeStr, ":")

		if len(keyValue) != 2 {
			continue
		}

		key := keyValue[0]
		value := keyValue[1]

		volumes = append(volumes, VolumeBinding{
			Container: key,
			Host:      value,
		})
	}
	return volumes
}

func mapToPorts(portMap nat.PortMap) []PortBinding {
	var ports []PortBinding
	for port, binds := range portMap {
		portNum := port.Port()
		protocol := port.Proto()

		if len(binds) == 0 {
			continue
		}

		bind := binds[0]
		hostPort := bind.HostPort

		ports = append(ports, PortBinding{
			Host:      hostPort,
			Container: portNum,
			Protocol:  protocol,
		})
	}
	return ports
}
