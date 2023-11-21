package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/aws"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/middleware"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
)

type JwtToken struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthController interface {
	CognitoCallback(code string) (*JwtToken, *error.RequestError)
	CognitoRefreshToken(token string) *error.RequestError
	CognitoGetCredentials(token string) (string, *error.RequestError)
}

func NewController(logger log.Logger, cognito aws.CognitoDriver) AuthController {
	return &controller{
		logger:  logger,
		cognito: cognito,
	}
}

type controller struct {
	cognito aws.CognitoDriver
	logger  log.Logger
}

func (c *controller) CognitoGetCredentials(token string) (string, *error.RequestError) {
	// Get the credentials from the token
	creds, err := c.cognito.GetCredentials(token)
	if err != nil {
		return "", error.NewRequestError(err, error.InternalServerError, "Error getting credentials", c.logger)
	}
	// Create JWT
	jwt, ok := middleware.SignToken(map[string]interface{}{
		"access_key_id":     *creds.AccessKeyId,
		"secret_access_key": *creds.SecretKey,
		"session_token":     *creds.SessionToken,
	})

	if !ok {
		return "", error.NewRequestError(nil, error.InternalServerError, "Error signing token", c.logger)
	}

	return jwt, nil
}

func (c *controller) CognitoRefreshToken(token string) *error.RequestError {
	return nil
}

func (c *controller) CognitoCallback(code string) (*JwtToken, *error.RequestError) {
	if code == "" {
		return nil, error.NewRequestError(nil, error.BadRequestError, "No code provided", c.logger)
	}

	clientID := config.EnvVars.Get(config.COGNITO_CLIENT_ID)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("redirect_uri", config.EnvVars.Get(config.COGNITO_REDIRECT_URL))

	// Get the token from the code
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", config.EnvVars.Get(config.COGNITO_AUTH_HOST)), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, error.NewRequestError(err, error.InternalServerError, "Error creating request", c.logger)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	basicAuth := fmt.Sprintf("%s:%s", clientID, config.EnvVars.Get(config.COGNITO_CLIENT_SECRET))
	// Base64 encode the basic auth
	encoded := base64.StdEncoding.EncodeToString([]byte(basicAuth))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encoded))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, error.NewRequestError(err, error.InternalServerError, "Error sending request", c.logger)
	}

	defer resp.Body.Close()

	// Get the token from the response
	token := JwtToken{}

	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, error.NewRequestError(err, error.InternalServerError, "Error decoding response", c.logger)
	}

	return &token, nil
}
