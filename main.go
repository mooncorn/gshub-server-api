package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/mooncorn/gshub-core/db"
	coreMiddlewares "github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/config"
	"github.com/mooncorn/gshub-server-api/handlers"
	"github.com/mooncorn/gshub-server-api/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	// Setup database
	gormDB := db.NewGormDB(config.Env.DSN)
	db.SetDatabase(gormDB)

	// Migrate the models
	err := db.GetDatabase().GetDB().AutoMigrate(
		&models.User{},
		&models.Plan{},
		&models.Service{},
		&models.ServiceEnv{},
		&models.ServiceEnvValue{},
		&models.ServiceVolume{},
		&models.ServicePort{},
		&models.Server{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if strings.ToLower(config.Env.GinMode) == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		ExposeHeaders: []string{"Content-Length"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"version": config.Env.Version})
	})

	r.Use(coreMiddlewares.CheckUser)
	r.Use(coreMiddlewares.RequireUser)
	r.Use(middlewares.CheckOwnership)

	r.GET("/state", handlers.GetState)
	r.GET("/console", handlers.GetConsole)
	r.POST("/run", handlers.RunCommand)
	r.GET("/env", handlers.GetEnv)

	r.POST("/start", handlers.StartServer)
	r.POST("/stop", handlers.StopServer)
	r.POST("/create", handlers.CreateServer)

	r.Run(":" + config.Env.Port)
}
