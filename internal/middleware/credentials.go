package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	api_error "github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	"github.com/golang-jwt/jwt/v5"
)

type Credentials struct {
	AccessKey    string `json:"access_key_id"`
	SecretKey    string `json:"secret_access_key"`
	SessionToken string `json:"session_token"`
}

func CognitorCredentialsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the credentials token from the cookies/headers
		credsToken, err := extractCredentials(w, r)
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.BadRequestError, "invalid credentials", nil))
			return
		}

		// Validate token
		creds, err := validateCredentials(credsToken)
		if err != nil {
			api_error.HandleError(w, r, api_error.NewRequestError(err, api_error.UnauthorizedError, "invalid credentials", nil))
			return
		}

		// Set the credentials in the context
		ctx := context.WithValue(r.Context(), CredentialsKey, creds)
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	})
}

func validateCredentials(token string) (*Credentials, error) {
	validToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("Unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.EnvVars.GetJWTPrivateKey()), nil
	})
	if err != nil {
		return nil, err
	}
	claims := validToken.Claims.(jwt.MapClaims)
	creds := &Credentials{
		AccessKey:    claims["access_key_id"].(string),
		SecretKey:    claims["secret_access_key"].(string),
		SessionToken: claims["session_token"].(string),
	}
	return creds, nil
}

func extractCredentials(w http.ResponseWriter, r *http.Request) (string, error) {
	// Check if the token is in the headers
	token := r.Header.Get("x-credentials")
	if token != "" {
		return token, nil
	}

	// Get the token from the cookies
	cookie, err := r.Cookie("credentials")
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
