package server

import (
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/auth"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

// Build creates a new router and adds the routes to it
func Build(logger log.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// Routes
	r.Get("/", rootRoute)

	// Handlers
	r.Mount("/auth", auth.Routes())
	r.Mount("/s3", s3.Routes(s3.NewController(logger, config.NewAwsDriver(logger))))

	// Print routes
	printEstablishedRoutes(r, logger)
	return r
}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	message := struct {
		Message string `json:"message"`
	}{Message: "Hello World"}
	render.JSON(w, r, message)
}

func printEstablishedRoutes(r *chi.Mux, logger log.Logger) {
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info("Route Established: ", method, " - ", route)
		return nil
	}
	if err := chi.Walk(r, walkFunc); err != nil {
		logger.Error("Error while printing routes: ", err.Error())
	}
}
