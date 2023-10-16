package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/api/idtoken"
)

// BuildAuthRouteHandler builds the routes for the auth handler
// and adds them to the given router
func Routes(r *chi.Mux, logger *config.Logger) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/google/sso", googleCallback)
	})
}

// googleCallback handles the callback from google
func googleCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract this to controller
	credential := r.FormValue("credential")
	if credential == "" {
		// TODO: Handle error
		fmt.Println("credential is empty")
		return
	}

	csrfToken := r.FormValue("g_csrf_token")
	if csrfToken == "" {
		// TODO: Handle error
		fmt.Println("csrfToken is empty")
		return
	}

	// Decdode credential
	// Get the google public key
	payload, err := idtoken.Validate(context.Background(), credential, config.EnvVars.Get(config.GOOGLE_CLIENT_ID))
	if err != nil {
		// TODO Handle error
		fmt.Println("error decoding credential")
		fmt.Println(err)
		return
	}
	claims := payload.Claims
	fmt.Print(payload.Claims)
	fmt.Println(credential)

	// Build user
	email, found := claims["email"].(string)
	if !found {
		// TODO Handle error
		fmt.Println("error getting email from payload")
		return
	}

	emailVerified, found := claims["email_verified"].(bool)
	if !found {
		// TODO Handle error
		fmt.Println("error getting email_verified from claims")
		return
	}

	name, found := claims["name"].(string)
	if !found {
		// TODO Handle error
		fmt.Println("error getting name from claims")
		return
	}

	picture, found := claims["picture"].(string)
	if !found {
		// TODO Handle error
		fmt.Println("error getting picture from claims")
		return
	}

	givenName, found := claims["given_name"].(string)
	if !found {
		// TODO Handle error
		fmt.Println("error getting given_name from claims")
		return
	}

	familyName, found := claims["family_name"].(string)
	if !found {
		// TODO Handle error
		fmt.Println("error getting family_name from claims")
		return
	}

	user := &types.User{
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
		GivenName:     givenName,
		FamilyName:    familyName,
	}

	fmt.Println(user)

	// Create jwt
	token, err := jwt.NewBuilder().Audience([]string{"family-cloud-api"}).Claim("user", user).Expiration(time.Now().Add(time.Hour * 24 * 7)).Build()
	if err != nil {
		// TODO Handle error
		fmt.Println("error creating jwt")
		fmt.Println(err)
		return
	}

	// Sign Token
	key, err := jwk.FromRaw([]byte(`secretkey`))
	if err != nil {
		fmt.Printf(`failed to create new symmetric key: %s`, err)
		return
	}
	key.Set(jws.KeyIDKey, `secret-key`)
	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, key))
	if err != nil {
		// TODO Handle error
		fmt.Println("error signing jwt")
		fmt.Println(err)
		return
	}

	fmt.Println(string(signedToken))

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   string(signedToken),
		Expires: time.Now().Add(time.Hour * 24 * 7),
	})

	// Redirect to frontend
	http.Redirect(w, r, "http://localhost:4200/home?sso=success", http.StatusFound)
}
