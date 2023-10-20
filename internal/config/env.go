package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

const (
	// PORT specifies the port where the server will be listening
	// Required for dev, and prod
	PORT = "PORT"

	// ENV specifies the environment where the server will be running
	// Required for dev, and prod
	ENV = "ENV"

	// DB_URI specifies the URI for the database
	// Required for dev, and prod
	DB_URI = "DB_URI"

	// GOOGLE_CLIENT_ID specifies the client id for the google oauth2
	// Required for dev, and prod
	GOOGLE_CLIENT_ID = "GOOGLE_CLIENT_ID"

	// GOOGLE_CLIENT_SECRET specifies the client secret for the google oauth2
	// Required for dev, and prod
	GOOGLE_CLIENT_SECRET = "GOOGLE_CLIENT_SECRET"

	// GOOGLE_REDIRECT_URL specifies the redirect url for the google oauth2
	// Required for dev, and prod
	GOOGLE_REDIRECT_URL = "GOOGLE_REDIRECT_URL"

	// GOOGLE_STATE specifies the state for the google oauth2
	// Required for dev, and prod
	GOOGLE_STATE = "GOOGLE_STATE"
)

var (
	// EnvVars is the struct holding the environment variables
	EnvVars *EnvConfig = newEnvConfig()

	// Required is the list of required environment variables for all environments
	required = []string{PORT, ENV, DB_URI, GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL}

	// DevRequired is the list of required environment variables for the development environment
	devRequired = []string{}

	// ProdRequired is the list of required environment variables for the production environment
	prodRequired = []string{}
)

// EnvConfig is the configuration for the environment variables
type EnvConfig struct {
	env map[string]string
}

// newEnvConfig creates a new EnvConfig
func newEnvConfig() *EnvConfig {
	env, err := godotenv.Read()
	if err != nil {
		fmt.Println("No .env file found.")
	}

	config := &EnvConfig{
		env: env,
	}
	config.validate()

	return config
}

// validate checks if the environment variables are set
func (e *EnvConfig) validate() {
	// Validate required environment variables
	e.validateRequired(required)
	environment := e.GetEnv()

	// Validate environment specific environment variables
	switch environment {
	case "development":
		e.validateRequired(devRequired)
	case "production":
		e.validateRequired(prodRequired)
	default:
		panic(fmt.Sprintf("Invalid environment: %s", environment))
	}
}

// validateRequired checks if the required environment variables are set
func (e *EnvConfig) validateRequired(required []string) {
	for _, key := range required {
		if e.env[key] == "" {
			panic(fmt.Sprintf("%s environment variable is not set", key))
		}
	}
}

// GetPort returns the port where the server will be listening
func (e *EnvConfig) GetPort() string {
	return e.env[PORT]
}

// GetEnv returns the environment where the server will be running
func (e *EnvConfig) GetEnv() string {
	return e.env[ENV]
}

// Get returns the value of a specific environment variable
func (e *EnvConfig) Get(key string) string {
	if e.env[key] == "" {
		panic(fmt.Sprintf("%s environment variable is not set", key))
	}

	return e.env[key]
}
