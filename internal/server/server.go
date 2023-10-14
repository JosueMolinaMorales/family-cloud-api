package server

import (
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/auth"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Build(logger *config.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/", rootRoute)

	// Handlers
	auth.BuildAuthRouteHandler(r, logger)
	return r
}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	message := struct {
		Message string `json:"message"`
	}{Message: "Hello World"}
	render.JSON(w, r, message)
}
