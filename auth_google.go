package pp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleUserinfo struct {
	Email string
}

type AuthGoogle struct {
	oauth *oauth2.Config
}

func NewAuthGoogle(clientID, clientSecret, authURL string) AuthGoogle {
	return AuthGoogle{&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  authURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}}
}

func (a AuthGoogle) HandleAuth(w http.ResponseWriter, r *http.Request, storage Storage, noSecureCookie bool) error {
	q := r.URL.Query()
	code := q.Get("code")

	if code == "" {
		log.Printf("auth code is not set, redirecting user to authentication")
		url := a.oauth.AuthCodeURL("state")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return nil
	}

	log.Print("exchanging auth code")

	tok, err := a.oauth.Exchange(r.Context(), code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %v", err)
	}

	client := a.oauth.Client(r.Context(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return fmt.Errorf("failed to get email: %v", err)
	}
	defer resp.Body.Close()

	jDecoder := json.NewDecoder(resp.Body)
	userinfo := googleUserinfo{}
	err = jDecoder.Decode(&userinfo)
	if err != nil {
		return fmt.Errorf("failed to parse JSON returned by Google: %v", err)
	}

	email := userinfo.Email
	secret, err := storage.CreateUser(email)
	if err != nil {
		return fmt.Errorf("failed to create a user: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "podcast_session",
		Value:    secret,
		Secure:   !noSecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	return nil
}
