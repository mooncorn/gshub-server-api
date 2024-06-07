package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Env Environment

type Environment struct {
	GinMode    string
	DSN        string
	URL        string
	Port       string
	JWTSecret  string
	InstanceId string
	Version    string
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

	Env = Environment{
		GinMode:    os.Getenv("GIN_MODE"),
		DSN:        os.Getenv("DSN"),
		URL:        os.Getenv("URL"),
		Port:       os.Getenv("PORT"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		InstanceId: os.Getenv("INSTANCE_ID"),
		Version:    os.Getenv("VERSION"),
	}
}
