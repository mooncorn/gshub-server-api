package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-server-api/app"
)

func GetConsole(c *gin.Context, appCtx *app.Context) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create docker client"})
		return
	}
	defer apiClient.Close()

	out, err := apiClient.ContainerLogs(c, "main", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get console logs stream", "details": err.Error()})
		return
	}

	logs, err := io.ReadAll(out)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read console logs", "details": err.Error()})
		return
	}

	// Remove padding
	cleanLogs := []string{}
	logsArray := strings.Split(string(logs), "\n")
	for _, line := range logsArray {
		if len(line) <= 8 {
			continue
		}
		cleanLogs = append(cleanLogs, line[8:])
	}

	c.JSON(http.StatusOK, gin.H{"console": cleanLogs})

}
