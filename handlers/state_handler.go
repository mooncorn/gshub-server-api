package handlers

import (
	"net/http"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

type Container struct {
	Id string `json:"id"`
}

func GetState(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"state": container.State.Status})
}
