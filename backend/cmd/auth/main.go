package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"

	_ "modernc.org/sqlite"
	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/server"
)

func generateState() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func generateDiscordUrl(clientId string, redirectURI string, state string) string {
	const baseURL string = "https://discord.com/oauth2/authorize"
	
	params := url.Values{
		"client_id": {clientId},
		"response_type": {"code"},
		"redirect_uri": {redirectURI},
		"scope": {"identify"},
		"state": {state},
	}

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func discord(clientId string, redirectURI string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()

		cookie := http.Cookie{
			Name: "oauthState",
			Value: state,
			Path: "/",
			MaxAge: 300, // 5 minutes
			HttpOnly: true,
			Secure: true,
			SameSite: http.SameSiteLaxMode, // needs to be lax so when user arrives back on website from discord, the cookie still persists
		}

		discordUrl := generateDiscordUrl(clientId, redirectURI, state)

		http.SetCookie(w, &cookie)

		http.Redirect(w, r, discordUrl, http.StatusFound)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("oauthState")
	if err != nil {
		switch {
			case errors.Is(err, http.ErrNoCookie):
           		http.Error(w, "cookie not found", http.StatusBadRequest)
        	default:
            	http.Error(w, "server error", http.StatusInternalServerError)
				panic(err)
		}
		return
	}

	state := r.URL.Query().Get("state")
	if state != cookie.Value {
		http.Error(w, "cookie state and query state do not match", http.StatusBadRequest)
		return
	}

	w.Write([]byte("Success!"))
}

func main() {
	s := server.Init()

	const DB_PATH = "/db/app.db"
	db, err := sql.Open("sqlite", DB_PATH)
	if err != nil {
		s.LogFatal("Failed to open sqlite database", "err", err)
	}
	defer db.Close()

	discordClientId := shared.Ensureenv("DISCORD_CLIENT_ID")
	redirectURI := shared.Ensureenv("REDIRECT_URI")

	s.Handle("/discord", s.BaseChain.ThenFunc(discord(discordClientId, redirectURI)))
	s.Handle("/redirect", s.BaseChain.ThenFunc(redirect))

	s.Listen(3002)
}