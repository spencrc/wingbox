package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type TokenRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserRes struct {
	UserId string `json:"id"`
}

var ErrInvalidState error = errors.New("the provided state code is invalid") 
var ErrMissingCode error = errors.New("code is missing from query parameters")

// Gets oauth state from cookie, checks its validity, then gets the code to obtain Discord access tokens.
// On failure, returns empty string and error.
// On success, returns code string and nil.
func redeemCodeFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return "", err
	}

	state := r.URL.Query().Get("state")
	if state != cookie.Value {
		return "", ErrInvalidState
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return "", ErrMissingCode
	}

	return code, nil
}

// Sends the request passed, checks if it responded OK, then decodes (with result put into passed data argument). Returns error.
// Due to how decoding works, data must be passed as a pointer!
func fetch[T TokenRes | UserRes](client *http.Client, req *http.Request, data *T) error {
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("api returned unexpected status %d", res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		return err
	}

	return nil
}

// Builds request to obtain Discord access token, then fetches a response
// On failure, returns empty TokenRes and error.
// On success, returns decoded response as TokenRes and nil.
func fetchTokenData(code string, redirectURI string, clientId string, clientSecret string, client *http.Client) (TokenRes, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	body.Set("client_id", clientId)
	body.Set("client_secret", clientSecret)

	req, err := http.NewRequest(http.MethodPost, DISCORD_BASE_URL + "/api/oauth2/token", strings.NewReader(body.Encode()))
	if err != nil {
		return TokenRes{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var tokenData TokenRes
	if err = fetch(client, req, &tokenData); err != nil {
		return TokenRes{}, err
	}

	return tokenData, nil
}

// Builds request to obtain current Discord user data, then fetches a response.
// On failure, returns empty UserRes and error.
// On success, returns decoded response as UserRes and nil.
func fetchDiscordUserData(tokenData TokenRes, client *http.Client) (UserRes, error) {
	req, err := http.NewRequest(http.MethodGet, DISCORD_BASE_URL + "/api/users/@me", nil)
	if err != nil {
		return UserRes{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

	var userData UserRes
	err = fetch(client, req, &userData)
	if err != nil {
		return UserRes{}, err
	}

	return userData, err
}

// Inserts into the database by Discord ID, and, if it returns no rows (due to Discord ID being unqiue), then finds row with matching Discord ID. Returns app's user ID.
// Could be done in one query by upserting instead; however, my understanding is that it will bottleneck concurrency by write locking when it should not
func ensureUser(db *sql.DB, discordID string, userID *uint64) error {
	const query = `
		INSERT INTO users (discord_id)
		VALUES (?)
		ON CONFLICT(discord_id) DO NOTHING
		RETURNING id;
	`
	insertErr := db.QueryRow(query, discordID).Scan(userID)
	switch insertErr {
	case sql.ErrNoRows:
		if selectErr := db.QueryRow("SELECT id FROM users WHERE discord_id =?", discordID).Scan(&userID); selectErr != nil {
			return selectErr
		}
	default:
		return insertErr
	}
	return nil
}

func (as *AuthService) Redirect(w http.ResponseWriter, r *http.Request) {
	logger := as.server.Logger
	db := as.server.Db
	client := &http.Client{}

	code, err := redeemCodeFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenData, err := fetchTokenData(code, as.redirectURI, as.clientId, as.clientSecret, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not fetch token from Discord", "err", err)
		return
	}

	discordUserData, err := fetchDiscordUserData(tokenData, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to fetch user data from Discord", "err", err)
		return
	}

	var userID uint64
	if err = ensureUser(db, discordUserData.UserId, &userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to insert or find user into database", "err", err)
		return
	}

	accessJti := uuid.NewString()
	accessClaims := map[string]string{
		"jti": accessJti,
		"sub": strconv.FormatUint(userID, 10),
	}

	refreshJti := uuid.NewString()
	refreshClaims := map[string]string{
		"jti": refreshJti,
		"sub": strconv.FormatUint(userID, 10),
	}

	err = as.accessMgr.SetJWTCookie(w, r, accessClaims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to set access token", "err", err)
		return
	}

	err = as.refreshMgr.SetJWTCookie(w, r, refreshClaims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to set refresh token", "err", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Access and refresh cookies set successfully"))
}