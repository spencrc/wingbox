package auth

import (
	"fmt"
	"net/http"
	"net/url"
)

// Generates OAuth URL for Discord
func generateDiscordUrl(clientId string, redirectURI string, state string) string {
	params := url.Values{
		"client_id":     {clientId},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"scope":         {"identify"},
		"state":         {state},
	}

	return fmt.Sprintf("%s?%s", DISCORD_BASE_URL + "/oauth2/authorize", params.Encode())
}

func (as *AuthService) Discord(w http.ResponseWriter, r *http.Request) {
	state := generateState()

	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // needs to be lax so when user arrives back on website from discord, the cookie still persists
	}

	discordUrl := generateDiscordUrl(as.clientId, as.redirectURI, state)

	http.SetCookie(w, &cookie)

	http.Redirect(w, r, discordUrl, http.StatusFound)
}