package handlers

import (
	"net/http"
	"strings"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	coreConfig "github.com/mooncorn/gshub-core/config"
	"github.com/mooncorn/gshub-server-api/app"
)

func GetEnv(c *gin.Context, appCtx *app.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	container, err := apiClient.ContainerInspect(c, "main")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get container info", "details": err.Error()})
		return
	}

	serviceConfig, err := coreConfig.GetServiceConfiguration(appCtx.ServiceData.ServiceNameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service configuration", "details": err.Error()})
		return
	}

	containerEnvArray := container.Config.Env
	// Convert to a map
	containerEnvMap := make(map[string]string)
	for _, containerEnv := range containerEnvArray {
		keyValue := strings.Split(containerEnv, "=")
		containerEnvMap[keyValue[0]] = keyValue[1]
	}

	// Filter out unwanted env vars
	values := make(map[string]string)
	for _, env := range serviceConfig.Env {
		value, ok := containerEnvMap[env.Key]
		if !ok {
			continue
		}

		values[env.Key] = value
	}

	c.JSON(http.StatusOK, gin.H{"values": values})
}
