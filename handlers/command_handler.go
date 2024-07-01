package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-server-api/app"
)

func RunCommand(c *gin.Context, appCtx *app.Context) {
	var request struct {
		Cmd string `json:"cmd"`
	}

	// Bind JSON input to the request structure
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Docker client"})
		return
	}

	formattedCmd, err := appCtx.ServiceController.FormatGameCommand(request.Cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature not supported"})
		return
	}

	// Command to execute
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash", "-c", formattedCmd},
	}

	execID, err := cli.ContainerExecCreate(c, "main", execConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exec instance"})
		return
	}

	// Attach to exec instance
	resp, err := cli.ContainerExecAttach(c, execID.ID, types.ExecStartCheck{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach to exec instance"})
		return
	}
	defer resp.Close()

	// Read output from exec instance
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read exec output"})
		return
	}

	cleanOutput := ""

	// Remove padding
	if len(string(output)) >= 8 {
		cleanOutput = string(output)[8:]
	}

	outputArray := strings.Split(cleanOutput, "\n")
	outputArray = outputArray[:len(outputArray)-1]

	c.JSON(http.StatusOK, gin.H{"output": outputArray})
}
