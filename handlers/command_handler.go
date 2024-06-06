package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/config"
	"github.com/mooncorn/gshub-server-api/core"
)

func RunCommand(c *gin.Context) {
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

	// Format the game command
	gameServer, err := core.NewGameServer(&service, &plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server configuration", "details": err.Error()})
		return
	}

	formattedCmd, err := gameServer.FormatGameCommand(request.Cmd)
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
