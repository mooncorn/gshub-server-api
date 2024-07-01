package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/utils"
	"github.com/mooncorn/gshub-server-api/app"
)

func CheckOwnership(appCtx *app.Context) func(c *gin.Context) {
	return func(c *gin.Context) {
		userIDStr := c.GetString("userID")
		userEmail := c.GetString("userEmail")

		userID64, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			utils.HandleError(c, http.StatusBadRequest, "Invalid user id", err, userEmail)
			return
		}

		if uint(userID64) != appCtx.ServiceData.OwnerID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
