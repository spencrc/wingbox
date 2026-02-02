package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	jwtcookie "github.com/stfsy/go-jwt-cookie"
	_ "modernc.org/sqlite"
	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/server"
)

type discordAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type discordCurrentUserResponse struct {
	UserId string `json:"id"`
}

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
		"client_id":     {clientId},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"scope":         {"identify"},
		"state":         {state},
	}

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func discord(clientId string, redirectURI string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()

		cookie := http.Cookie{
			Name:     "oauthState",
			Value:    state,
			Path:     "/",
			MaxAge:   300, // 5 minutes
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode, // needs to be lax so when user arrives back on website from discord, the cookie still persists
		}

		discordUrl := generateDiscordUrl(clientId, redirectURI, state)

		http.SetCookie(w, &cookie)

		http.Redirect(w, r, discordUrl, http.StatusFound)
	}
}

func redirect(redirectURI string, discordClientId string, discordClientSecret string, logger *slog.Logger, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "query code not found", http.StatusBadRequest)
			return
		}

		body := url.Values{}
		body.Set("grant_type", "authorization_code")
		body.Set("code", code)
		body.Set("redirect_uri", redirectURI)
		body.Set("client_id", discordClientId)
		body.Set("client_secret", discordClientSecret)

		req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/oauth2/token", strings.NewReader(body.Encode()))
		if err != nil {
			http.Error(w, "could not create request", http.StatusInternalServerError)
			logger.Warn("could not create request", "err", err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		tokenRes, err := client.Do(req)
		if err != nil {
			http.Error(w, "could not make request to Discord", http.StatusInternalServerError)
			logger.Warn("could not make request to Discord", "err", err)
			return
		}
		defer tokenRes.Body.Close()

		if tokenRes.StatusCode != http.StatusOK {
			logger.Warn("discord rejected token exchange", "status", tokenRes.StatusCode)
			http.Error(w, "Discord rejected the token exchange", tokenRes.StatusCode)
			return
		}

		var tokenData discordAccessTokenResponse
		err = json.NewDecoder(tokenRes.Body).Decode(&tokenData)
		if err != nil {
			http.Error(w, "could not decode JSON response", http.StatusInternalServerError)
			logger.Warn("could not decode JSON response", "err", err)
			return
		}

		req, err = http.NewRequest(http.MethodGet, "https://discord.com/api/users/@me", nil)
		if err != nil {
			http.Error(w, "could not create request", http.StatusInternalServerError)
			logger.Warn("could not create request", "err", err)
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

		userRes, err := client.Do(req)
		if err != nil {
			http.Error(w, "could not make request to Discord", http.StatusInternalServerError)
			logger.Warn("could not make request to Discord", "err", err)
			return
		}
		defer userRes.Body.Close()

		if userRes.StatusCode != http.StatusOK {
			// This will tell you exactly what went wrong (e.g., 400 Bad Request)
			logger.Warn("discord rejected current user request", "status", userRes.StatusCode)
			http.Error(w, "Discord rejected the token exchange", userRes.StatusCode)
			return
		}

		var discordUserData discordCurrentUserResponse
		err = json.NewDecoder(userRes.Body).Decode(&discordUserData)
		if err != nil {
			http.Error(w, "could not decode JSON response", http.StatusInternalServerError)
			logger.Warn("could not decode JSON response", "err", err)
			return
		}

		w.Write([]byte(discordUserData.UserId))
		query := `
			INSERT INTO users (discord_id)
			VALUES (?)
			ON CONFLICT(discord_id) DO UPDATE SET discord_id = excluded.discord_id
			RETURNING id
		`
		// now needs to upset into sqlite DB, then grant tokens
		var userId uint64
		err = db.QueryRow(query, discordUserData.UserId).Scan(&userId)
		if err != nil {
			http.Error(w, "could not add user to database", http.StatusInternalServerError)
			logger.Warn("could not add user to database", "err", err, "discord_id", discordUserData.UserId)
			return
		}

		jwtSecret := shared.Ensureenv("JWT_SECRET")
		manager, err := jwtcookie.NewCookieManager(
			jwtcookie.WithHTTPOnly(true),
			jwtcookie.WithSecure(true),
			jwtcookie.WithSigningKeyHMAC(
				[]byte(jwtSecret),
				[]byte("0123456789abcdef"), // 16-byte salt example; prefer 32 random bytes in production
			),
			jwtcookie.WithSigningMethod(jwt.SigningMethodHS256),
			//TODO: add issuer and audience domains! more secure
		)

		accessToken := map[string]string{
			"user_id": strconv.FormatUint(userId, 10),
		}

		//sign access token
	}
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
	discordClientSecret := shared.Ensureenv("DISCORD_CLIENT_SECRET")
	redirectURI := shared.Ensureenv("REDIRECT_URI")

	s.Handle("/discord", s.BaseChain.ThenFunc(discord(discordClientId, redirectURI)))
	s.Handle("/redirect", s.BaseChain.ThenFunc(redirect(redirectURI, discordClientId, discordClientSecret, s.Logger, db)))

	s.Listen(3002)
}
