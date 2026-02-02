package auth

import (
	"fmt"
	"math/rand"
	"net/url"
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

// Generates OAuth URL for Discord
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