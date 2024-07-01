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
	"github.com/mooncorn/gshub-core/config"
	coreConfig "github.com/mooncorn/gshub-core/config"
	"github.com/mooncorn/gshub-server-api/app"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func StartServer(c *gin.Context, appCtx *app.Context) {
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

func StopServer(c *gin.Context, appCtx *app.Context) {
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

func CreateService(c *gin.Context, appCtx *app.Context) {
	var request struct {
		Config map[string]string `json:"config"`
		Type   string            `json:"type"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	config, err := appCtx.ServiceController.ValidateConfig(request.Env)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server configuration", "details": err.Error()})
		return
	}

	serviceConfig, err := coreConfig.GetServiceConfiguration(appCtx.ServiceData.ServiceNameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service configuration", "details": err.Error()})
		return
	}

	containerEnv := FormatEnv(config)
	containerPorts := FormatPorts(serviceConfig.Ports)
	containerVolumes := FormatVolumes(serviceConfig.Volumes)

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
			if tag == appCtx.ServiceData.ServiceImage {
				imageExists = true
				break
			}
		}
		if imageExists {
			break
		}
	}

	if !imageExists {
		out, err := apiClient.ImagePull(c, appCtx.ServiceData.ServiceImage, image.PullOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull image", "details": err.Error()})
			return
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}

	if _, err := apiClient.ContainerCreate(c, &container.Config{
		Env:   containerEnv,
		Image: appCtx.ServiceData.ServiceImage,
	}, &container.HostConfig{
		PortBindings: containerPorts,
		Binds:        containerVolumes,
	}, &network.NetworkingConfig{}, &v1.Platform{}, "main"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server", "details": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func UpdateServer(c *gin.Context, appCtx *app.Context) {}

func DeleteServer(c *gin.Context, appCtx *app.Context) {
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

func FormatPorts(ports []config.Port) map[nat.Port][]nat.PortBinding {
	portBindings := make(map[nat.Port][]nat.PortBinding)
	for _, port := range ports {
		containerPort := nat.Port(fmt.Sprintf("%d/%s", port.Container, port.Protocol))
		hostPort := nat.PortBinding{
			HostPort: fmt.Sprintf("%d", port.Host),
		}
		portBindings[containerPort] = append(portBindings[containerPort], hostPort)
	}
	return portBindings
}

func FormatVolumes(volumes []config.Volume) []string {
	binds := make([]string, len(volumes))

	for i, vol := range volumes {
		binds[i] = fmt.Sprintf("%s:%s", vol.Host, vol.Destination)
	}

	return binds
}
