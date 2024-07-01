package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	coreMiddlewares "github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-server-api/app"
	"github.com/mooncorn/gshub-server-api/config"
	"github.com/mooncorn/gshub-server-api/handlers"
	"github.com/mooncorn/gshub-server-api/middlewares"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()

	gormDB := initializeDatabase()

	// Get initialization data from main api
	appCtx := app.NewContext(gormDB)

	go monitorUptime(appCtx)

	if strings.ToLower(config.Env.AppEnv) == "production" {
		gin.SetMode(gin.ReleaseMode)
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
	r.Use(middlewares.CheckOwnership(appCtx))

	r.GET("/state", appCtx.HandlerWrapper(handlers.GetState))
	r.GET("/console", appCtx.HandlerWrapper(handlers.GetConsole))
	r.POST("/run", appCtx.HandlerWrapper(handlers.RunCommand))
	r.GET("/env", appCtx.HandlerWrapper(handlers.GetEnv))

	r.POST("/start", appCtx.HandlerWrapper(handlers.StartServer))
	r.POST("/stop", appCtx.HandlerWrapper(handlers.StopServer))
	r.POST("/create", appCtx.HandlerWrapper(handlers.CreateServer))
	r.DELETE("/remove", appCtx.HandlerWrapper(handlers.DeleteServer))

	server := &http.Server{
		Addr:    ":" + config.Env.Port,
		Handler: r,
	}

	// Run the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Set up channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-stop

	log.Println("Shutting down gracefully...")

	// Create a context with timeout for the shutdown
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send burned cycles to main api
	if err := cleanup(appCtx); err != nil {
		log.Fatalf("Cleanup failed: %v", err)
		// TODO: save burned cycles to local sqlite
	}

	log.Println("Uptime sent successfully")
	log.Println("Server exiting")
}

func cleanup(appCtx *app.Context) error {
	err := appCtx.CyclesApiClient.PostShutdown(appCtx.BurnedCycles)
	return err
}

func monitorUptime(appCtx *app.Context) {
	for {
		appCtx.BurnedCycles++

		fmt.Printf("Burned cycles: %d/%d\n", appCtx.BurnedCycles, appCtx.InstancePayload.Cycles)

		if appCtx.InstancePayload.Cycles <= appCtx.BurnedCycles {
			fmt.Println("Allowed uptime reached. Shutting down...")
			// appCtx.SystemController.Shutdown() // TODO: execute only in production
			p, _ := os.FindProcess(os.Getpid())
			p.Signal(syscall.SIGINT)
			return
		}

		time.Sleep(time.Second)
	}
}

func initializeDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	return db
}
