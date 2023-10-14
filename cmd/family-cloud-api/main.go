package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/server"
)

func main() {
	logger := config.NewLogger().With(context.TODO(), "Version", "1.0.0")
	server := server.Build(&logger)
	port := config.EnvVars.GetPort()

	http.ListenAndServe(fmt.Sprintf(":%s", port), server)
}
