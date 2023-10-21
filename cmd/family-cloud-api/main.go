package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/server"
)

func main() {
	logger := log.NewLogger().With(context.TODO(), "Version", "1.0.0")
	server := server.Build(logger)
	port := config.EnvVars.GetPort()

	logger.Info("Starting server on port ", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), server)
}
