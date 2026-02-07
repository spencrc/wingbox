package auth

import (
	"math/rand"
	"net/http"
)

// Generates state code for PKCE for Discord
func generateState() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func generateStateCookie(state string) http.Cookie {
	return http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // needs to be lax so when user arrives back on website from discord, the cookie still persists
	}
}