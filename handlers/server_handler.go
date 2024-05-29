package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/helpers"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const MEMORY_NAME = "Memory"

func StartServer(c *gin.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	err = apiClient.ContainerStart(c, "main", container.StartOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start server", "details": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func StopServer(c *gin.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	err = apiClient.ContainerStop(c, "main", container.StopOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start server", "details": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func CreateServer(c *gin.Context) {
	var request struct {
		Env map[string]string `json:"env"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	dbInstance := db.GetDatabase()

	// Get server based on INSTANCE_ID
	var server models.Server
	if err := dbInstance.GetDB().Where(&models.Server{InstanceID: helpers.INSTANCE_ID}).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Server not found"})
		return
	}

	// Get service
	var service models.Service
	if err := dbInstance.GetDB().Preload("Env.Values").Preload("Ports").Preload("Volumes").Where(&models.Service{ID: server.ServiceID}).First(&service).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Service not found"})
		return
	}

	// Get plan
	var plan models.Plan
	if err := dbInstance.GetDB().Where(&models.Plan{ID: server.PlanID}).First(&plan).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Plan not found"})
		return
	}

	containerEnv, err := ProcessEnvVars(request.Env, service, plan)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid server configuration", "details": err.Error()})
		return
	}

	containerPorts := FormatPorts(service.Ports)
	containerVolumes := FormatVolumes(service.Volumes)

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	// Check if the image exists and pull it if it does not
	images, err := apiClient.ImageList(c, types.ImageListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list images", "details": err.Error()})
		return
	}

	imageExists := false
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == service.Image {
				imageExists = true
				break
			}
		}
		if imageExists {
			break
		}
	}

	fmt.Println(imageExists)

	if !imageExists {
		out, err := apiClient.ImagePull(c, service.Image, image.PullOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull image", "details": err.Error()})
			return
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}

	if _, err := apiClient.ContainerCreate(c, &container.Config{
		Env:   containerEnv,
		Image: service.Image,
	}, &container.HostConfig{
		PortBindings: containerPorts,
		Binds:        containerVolumes,
	}, &network.NetworkingConfig{}, &v1.Platform{}, "main"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server", "details": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func UpdateServer(c *gin.Context) {}

func DeleteServer(c *gin.Context) {}

// Process environment variables for the given service and plan configuration.
func ProcessEnvVars(config map[string]string, service models.Service, plan models.Plan) ([]string, error) {
	var containerEnv []string
	var memoryEnv *models.ServiceEnv

	// Validate and process each environment variable from the service
	for _, env := range service.Env {
		if env.Name == MEMORY_NAME {
			memoryEnv = &env
			continue
		}

		value, ok := config[env.Key]
		if !ok {
			if env.Required {
				return nil, fmt.Errorf("%s is required", env.Key)
			}
			containerEnv = append(containerEnv, fmt.Sprintf("%s=%s", env.Key, env.Default))
			continue
		}

		if !IsValidEnvValue(env.Values, value) {
			return nil, fmt.Errorf("invalid %s: %s", env.Key, value)
		}

		containerEnv = append(containerEnv, fmt.Sprintf("%s=%s", env.Key, value))
	}

	// Process memory environment variable if it exists
	if memoryEnv != nil {
		if err := ProcessMemoryEnvVar(memoryEnv, service, plan, &containerEnv); err != nil {
			return nil, err
		}
	}

	return containerEnv, nil
}

// Check if the provided value is one of the allowed values.
func IsValidEnvValue(values []models.ServiceEnvValue, value string) bool {
	for _, envValue := range values {
		if envValue.Value == value {
			return true
		}
	}
	return false
}

// Process the memory-specific environment variable.
func ProcessMemoryEnvVar(memoryEnv *models.ServiceEnv, service models.Service, plan models.Plan, containerEnv *[]string) error {
	serverMemory := helpers.CalculateServerMemory(plan.Memory)
	if !helpers.CheckMinimumMemory(serverMemory, service.MinMem) {
		return errors.New("not enough memory for this service")
	}

	envValue, err := helpers.FindEnvMemoryValue(service, serverMemory, MEMORY_NAME)
	if err != nil {
		return errors.New("appropriate memory env value not found")
	}

	*containerEnv = append(*containerEnv, fmt.Sprintf("%s=%s", memoryEnv.Key, envValue.Value))
	return nil
}

func FormatPorts(ports []models.ServicePort) map[nat.Port][]nat.PortBinding {
	portBindings := make(map[nat.Port][]nat.PortBinding)
	for _, port := range ports {
		containerPort := nat.Port(fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol))
		hostPort := nat.PortBinding{
			HostPort: fmt.Sprintf("%d", port.HostPort),
		}
		portBindings[containerPort] = append(portBindings[containerPort], hostPort)
	}
	return portBindings
}

func FormatVolumes(volumes []models.ServiceVolume) []string {
	var binds []string
	for _, vol := range volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", vol.Host, vol.Destination))
	}
	return binds
}
