package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Env Environment

type Environment struct {
	AppEnv     string
	DSN        string
	URL        string
	Port       string
	JWTSecret  string
	InstanceId string
	// InstanceMemory               int
	// ServiceNameID                string
	// ServiceMinimumMemoryRequired int
	CyclesUrl string
	// OwnerID                      uint
}

func LoadEnv() {
	env := os.Getenv("APP_ENV")
	var envFile string

	switch env {
	case "production":
		envFile = ".env.production"
	case "development":
		envFile = ".env.development"
	default:
		log.Fatal("APP_ENV has to be set to \"production\" or \"development\"")
	}

	log.Printf("APP_ENV set to \"%s\"", env)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	// // convert instance memory to int
	// instanceMemoryStr := os.Getenv("INSTANCE_MEMORY")
	// instanceMemory, err := strconv.Atoi(instanceMemoryStr)
	// if err != nil {
	// 	log.Fatalf("invalid INSTANCE_MEMORY env value: %s", instanceMemoryStr)
	// }

	// // convert service minimum memory required to int
	// serviceMinimumMemoryRequiredStr := os.Getenv("SERVICE_MINIMUM_MEMORY_REQUIRED")
	// serviceMinimumMemoryRequired, err := strconv.Atoi(serviceMinimumMemoryRequiredStr)
	// if err != nil {
	// 	log.Fatalf("invalid SERVICE_MINIMUM_MEMORY_REQUIRED env value: %s", serviceMinimumMemoryRequiredStr)
	// }

	// // convert owner id to uint
	// ownerIDStr := os.Getenv("OWNER_ID")
	// ownerID, err := strconv.ParseUint(ownerIDStr, 10, 32)
	// if err != nil {
	// 	log.Fatalf("invalid OWNER_ID env value: %s", ownerIDStr)
	// }

	Env = Environment{
		AppEnv:     os.Getenv("APP_ENV"),
		DSN:        os.Getenv("DSN"),
		URL:        os.Getenv("URL"),
		Port:       os.Getenv("PORT"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		InstanceId: os.Getenv("INSTANCE_ID"),
		// InstanceMemory:               instanceMemory,
		// ServiceNameID:                os.Getenv("SERVICE_NAME_ID"),
		// ServiceMinimumMemoryRequired: serviceMinimumMemoryRequired,
		CyclesUrl: os.Getenv("CYCLES_URL"),
		// OwnerID:                      uint(ownerID),
	}
}
