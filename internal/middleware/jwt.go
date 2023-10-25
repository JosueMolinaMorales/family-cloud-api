package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	api_error "github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// TokenKey is the key for the token in the context
	TokenKey ContextKey = "token"
)

// Token is the token after it has been parsed and validated
type Token struct {
	Groups        []string `json:"cognito:groups"`
	Username      string   `json:"cognito:username"`
	PreferredRole string   `json:"cognito:preferred_role"`
	Roles         []string `json:"cognito:roles"`
	ID            string   `json:"sub"`
	Email         string   `json:"email"`
}

// AuthMiddlware is the middleware for authenticating requests
func AuthMiddlware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		token, err := extractToken(w, r)
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.UnauthorizedError, "invalid token", nil))
			return
		}

		// Get the public key
		jwks, err := getPublicKey()
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.InternalServerError, "error getting public key", nil))
			return
		}

		validToken, err := validateToken(token, jwks)
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.UnauthorizedError, "invalid token", nil))
			return
		}

		tokenObj, err := getTokenObject(validToken)
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.InternalServerError, "error getting token object", nil))
			return
		}

		// Set the token in the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, TokenKey, tokenObj)
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}

func getTokenObject(validToken *jwt.Token) (*Token, error) {
	var tokenObj Token
	claims, ok := validToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	marshalledClaims, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(marshalledClaims, &tokenObj); err != nil {
		return nil, err
	}

	return &tokenObj, nil
}

func validateToken(token string, jwks *keyfunc.JWKS) (*jwt.Token, error) {
	validToken, err := jwt.Parse(token, jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return nil, err
	}
	if !validToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return validToken, nil
}

func getPublicKey() (*keyfunc.JWKS, error) {
	// Get the public key
	// TODO: Cache the public key
	jwks, err := keyfunc.Get(config.EnvVars.Get(config.COGNITO_JWKS_URL), keyfunc.Options{})
	if err != nil {
		return nil, err
	}
	return jwks, nil
}

func extractToken(w http.ResponseWriter, r *http.Request) (string, error) {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid authorization header")
		}
		return parts[1], nil
	}
	// Check if the token is in the cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", fmt.Errorf("no token found")
	}

	return cookie.Value, nil
}
