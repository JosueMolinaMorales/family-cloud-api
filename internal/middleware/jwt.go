package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
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
		var token string

		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) < 2 {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Println("No token")
				return
			}
			token = parts[1]
		} else {
			// Check if the token is in the cookie
			if cookie, err := r.Cookie("token"); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Println("No token")
				return
			} else {
				token = cookie.Value
			}
		}

		// Get the public key
		// TODO: Cache the public key
		jwks, err := keyfunc.Get(config.EnvVars.Get(config.COGNITO_JWKS_URL), keyfunc.Options{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Get the public key
		validToken, err := jwt.Parse(token, jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err.Error())
			return
		}
		if !validToken.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err.Error())
			return
		}

		// Set the token in the context
		var tokenObj Token
		claims, ok := validToken.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		marshalledClaims, err := json.Marshal(claims)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(marshalledClaims, &tokenObj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, TokenKey, tokenObj)
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}
