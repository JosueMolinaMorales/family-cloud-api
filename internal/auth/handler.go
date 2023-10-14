package auth

import (
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/go-chi/chi/v5"
)

// BuildAuthRouteHandler builds the routes for the auth handler
// and adds them to the given router
func BuildAuthRouteHandler(r *chi.Mux, logger *config.Logger) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", login)
		r.Post("/register", register)
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login"))
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("register"))
}
