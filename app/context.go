package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-server-api/config"
	"github.com/mooncorn/gshub-server-api/internal"
	"github.com/mooncorn/gshub-server-api/service"
	"github.com/mooncorn/gshub-server-api/system"
	"gorm.io/gorm"
)

type Context struct {
	DB                *gorm.DB
	BurnedCycles      uint
	StartupPayload    *internal.StartupPayload
	ServiceController *service.ServiceController
	SystemController  *system.AmazonLinuxSystemController
	CyclesApiClient   *internal.ApiClient
}

func NewContext(dbInstance *gorm.DB) *Context {
	client := internal.NewClient()

	// get failed burned cycles total amount if any
	var sum uint
	if err := dbInstance.Find(&internal.FailedBurnedCycle{}).Select("SUM(amount)").Row().Scan(&sum); err != nil {
		log.Fatalf("failed to get failed burned cycles total amount")
	}

	fmt.Printf("Failed burned cycles total amount: %d", sum)

	// fetch init data
	startupPayload, err := client.PostStartup(sum)
	if err != nil {
		log.Fatalf("failed to fetch service data: %v", err)
	}

	// delete all failed burned cycles
	// if this fails both the instance and main api state will be unsynchronized
	// meaning that on the next startup there will be a duplication of failed burned cycles
	if err = dbInstance.Where("1 = 1").Delete(&internal.FailedBurnedCycle{}).Error; err != nil {
		log.Fatalf("failed to delete all failed burned cycles: %v", err)
	}

	serviceController, err := service.NewServiceController(&service.InstanceData{
		StartupPayload: *startupPayload,
		InstanceID:     config.Env.InstanceId,
	})
	if err != nil {
		log.Fatalf("failed to create the service controller: %v", err)
	}

	return &Context{
		DB:                dbInstance,
		BurnedCycles:      0,
		StartupPayload:    startupPayload,
		ServiceController: serviceController,
		SystemController:  system.NewAmazonLinuxSystemController(),
		CyclesApiClient:   client,
	}
}

func (appCtx *Context) HandlerWrapper(handler func(*gin.Context, *Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, appCtx)
	}
}
