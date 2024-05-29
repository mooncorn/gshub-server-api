package main

import (
	"log"
	"os"

	"github.com/mooncorn/gshub-core/db"
	coreMiddlewares "github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-server-api/handlers"
	"github.com/mooncorn/gshub-server-api/helpers"
	"github.com/mooncorn/gshub-server-api/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve instance ID
	helpers.GetInstanceId()

	port := os.Getenv("PORT")
	dsn := os.Getenv("DSN")

	gormDB := db.NewGormDB(dsn)
	db.SetDatabase(gormDB)

	// AutoMigrate the models
	err = db.GetDatabase().GetDB().AutoMigrate(&models.Plan{}, &models.Service{}, &models.Server{}, &models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := gin.Default()

	// Middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		ExposeHeaders: []string{"Content-Length"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	r.Use(coreMiddlewares.CheckUser)
	r.Use(coreMiddlewares.RequireUser)
	r.Use(middlewares.CheckOwnership)

	r.GET("/state", handlers.GetState)
	r.GET("/console", handlers.GetConsole)
	r.POST("/run", handlers.RunCommand)
	r.GET("/env", handlers.GetEnv)

	r.POST("/servers", handlers.CreateServer)

	r.Run(":" + port)
}
