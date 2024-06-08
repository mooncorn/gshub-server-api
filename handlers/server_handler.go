package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/config"
	"github.com/mooncorn/gshub-server-api/core"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

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
	if err := dbInstance.GetDB().Where(&models.Server{InstanceID: config.Env.InstanceId}).First(&server).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server not found"})
		return
	}

	// Get service
	var service models.Service
	if err := dbInstance.GetDB().Preload("Env.Values").Preload("Ports").Preload("Volumes").Where(&models.Service{ID: server.ServiceID}).First(&service).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service not found"})
		return
	}

	// Get plan
	var plan models.Plan
	if err := dbInstance.GetDB().Where(&models.Plan{ID: server.PlanID}).First(&plan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan not found"})
		return
	}

	gameServer, err := core.NewGameServer(&service, &plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server configuration", "details": err.Error()})
		return
	}

	config, err := gameServer.ValidateConfig(request.Env)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server configuration", "details": err.Error()})
		return
	}

	containerEnv := FormatEnv(config)
	containerPorts := FormatPorts(service.Ports)
	containerVolumes := FormatVolumes(service.Volumes)

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	// Check if the image exists and pull it if it does not
	images, err := apiClient.ImageList(c, image.ListOptions{})
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

func DeleteServer(c *gin.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	if err := apiClient.ContainerRemove(c, "main", container.RemoveOptions{
		RemoveVolumes: false,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove server", "details": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func FormatEnv(env map[string]string) []string {
	formattedEnv := make([]string, 0, len(env))
	for key, value := range env {
		formattedEnv = append(formattedEnv, key+"="+value)
	}
	return formattedEnv
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
	binds := make([]string, len(volumes))

	for i, vol := range volumes {
		binds[i] = fmt.Sprintf("%s:%s", vol.Host, vol.Destination)
	}

	return binds
}
