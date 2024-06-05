package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/config"
)

func CheckOwnership(c *gin.Context) {
	email := c.GetString("userEmail")

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
	if err := dbInstance.GetDB().Where(&models.Server{UserID: user.ID, InstanceID: config.Env.InstanceId}).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access unauthorized"})
		c.Abort()
		return
	}
}
