package auth

import "net/http"

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