package handlers

import (
	"net/http"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-server-api/app"
)

type Container struct {
	Id string `json:"id"`
}

func GetState(c *gin.Context, appCtx *app.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	container, err := apiClient.ContainerInspect(c, "main")
	if err != nil {

		if client.IsErrNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get server info", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": container.State.Status})
}
