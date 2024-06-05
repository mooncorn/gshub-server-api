package handlers

import (
	"net/http"
	"strings"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/config"
)

func GetEnv(c *gin.Context) {
	var server models.Server
	dbInstance := db.GetDatabase()
	if err := dbInstance.GetDB().Where(&models.Server{InstanceID: config.Env.InstanceId}).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Server not found"})
		return
	}

	var service models.Service
	if err := dbInstance.GetDB().Where(&models.Service{ID: server.ServiceID}).First(&service).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Service not found"})
		return
	}

	var envs []models.ServiceEnv
	if err := dbInstance.GetDB().Where(&models.ServiceEnv{ServiceID: service.ID}).Find(&envs).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Envs not found"})
		return
	}

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

	containerEnvArray := container.Config.Env
	// Convert to a map
	containerEnvMap := make(map[string]string)
	for _, containerEnv := range containerEnvArray {
		keyValue := strings.Split(containerEnv, "=")
		containerEnvMap[keyValue[0]] = keyValue[1]
	}

	// Filter out unwanted env vars
	values := make(map[string]string)
	for _, env := range envs {
		value, ok := containerEnvMap[env.Key]
		if !ok {
			continue
		}

		values[env.Key] = value
	}

	c.JSON(http.StatusOK, gin.H{"values": values})
}
