package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/helpers"
)

func CheckOwnership(c *gin.Context) {
	email := c.GetString("userEmail")

	instanceID, err := helpers.GetInstanceId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instance id", "details": err.Error()})
		c.Abort()
		return
	}

	dbInstance := db.GetDatabase()

	// Get User
	var user models.User
	if err := dbInstance.GetDB().Where(&models.User{Email: email}).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user"})
		c.Abort()
		return
	}

	// Check if this instance belongs to this user
	var server models.Server
	if err := dbInstance.GetDB().Where(&models.Server{UserID: user.ID, InstanceID: instanceID}).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access unauthorized"})
		c.Abort()
		return
	}
}
