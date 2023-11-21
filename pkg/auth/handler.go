package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/middleware"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Routes returns the routes for the auth package
func Routes(controller AuthController) *chi.Mux {
	router := chi.NewRouter()

	h := &handler{
		controller: controller,
		logger:     log.NewLogger().With(context.Background(), "Version", "1.0.0"),
	}

	router.Get("/cognito/callback", h.CognitoCallback)
	router.Get("/cognito/refreshtoken", h.CognitoRefreshToken)

	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlware)
		r.Get("/cognito/credentials", h.GenerateTemporaryCredentials)
	})
	return router
}

type handler struct {
	controller AuthController
	logger     log.Logger
}

func (h *handler) GenerateTemporaryCredentials(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(middleware.JWTKey).(string)
	if !ok {
		error.HandleError(w, r, error.NewRequestError(nil, error.InternalServerError, "Error getting token from context", h.logger))
		return
	}

	jwt, err := h.controller.CognitoGetCredentials(token)
	if err != nil {
		error.HandleError(w, r, err)
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:    "credentials",
		Value:   jwt,
		Expires: time.Now().Add(time.Hour),
		Path:    "/",
	})

	// Return the credentials
	render.JSON(w, r, map[string]interface{}{
		"credentials": jwt,
	})
}

func (h *handler) CognitoRefreshToken(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Get User from context

	jwt, ok := r.Context().Value(middleware.JWTKey).(string)
	if !ok {
		error.HandleError(w, r, error.NewRequestError(nil, error.InternalServerError, "Error getting token from context", h.logger))
		return
	}
	if err := h.controller.CognitoRefreshToken(jwt); err != nil {
		error.HandleError(w, r, err)
		return
	}
}

func (h *handler) CognitoCallback(w http.ResponseWriter, r *http.Request) {
	clientUrl := config.EnvVars.Get(config.CLIENT_URL)
	token, err := h.controller.CognitoCallback(r.URL.Query().Get("code"))
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/home?sso=error", clientUrl), http.StatusFound)
		error.HandleError(w, r, err)
		return
	}

	// Generate Temporary Credentials
	creds, err := h.controller.CognitoGetCredentials(token.IDToken)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/home?sso=error", clientUrl), http.StatusFound)
		error.HandleError(w, r, err)
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:    "credentials",
		Value:   creds,
		Expires: time.Now().Add(time.Hour),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token.IDToken,
		Expires: time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
		Path:    "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "refresh_token",
		Value: token.RefreshToken,
		Path:  "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   token.AccessToken,
		Expires: time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
		Path:    "/",
	})

	// Redirect to frontend
	http.Redirect(w, r, fmt.Sprintf("%s/home?sso=success", clientUrl), http.StatusFound)
}
