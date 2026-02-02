package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	jwtcookie "github.com/stfsy/go-jwt-cookie"
	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/server"
)

type AuthService struct {
	server *server.Server
	redirectURI string
	clientId string
	clientSecret string
	accessMgr *jwtcookie.CookieManager
	refreshMgr *jwtcookie.CookieManager
}

func newAccessManager(jwtKey []byte, jwtSalt []byte) (*jwtcookie.CookieManager, error) {
	return jwtcookie.NewCookieManager(
		jwtcookie.WithHTTPOnly(true),
		jwtcookie.WithSecure(true),
		jwtcookie.WithSigningKeyHMAC(
			jwtKey,
			jwtSalt,
		),
		jwtcookie.WithValidationKeysHMAC([][]byte{jwtKey}),
		jwtcookie.WithSigningMethod(jwt.SigningMethodHS256),
		jwtcookie.WithMaxAge(ACCESS_MAX_AGE),
		jwtcookie.WithIssuer("auth"),
		jwtcookie.WithAudience("wingbox"),
		jwtcookie.WithSameSite(http.SameSiteLaxMode),
		jwtcookie.WithCookieName("__Http-DO_NOT_SHARE-access_token"),
	)
}

func newRefreshManager(jwtKey []byte, jwtSalt []byte) (*jwtcookie.CookieManager, error) {
	return jwtcookie.NewCookieManager(
		jwtcookie.WithHTTPOnly(true),
		jwtcookie.WithSigningKeyHMAC(
			jwtKey,
			jwtSalt,
		),
		jwtcookie.WithValidationKeysHMAC([][]byte{jwtKey}),
		jwtcookie.WithSigningMethod(jwt.SigningMethodHS256),
		jwtcookie.WithMaxAge(REFRESH_MAX_AGE), // 1 month
		jwtcookie.WithIssuer("auth"),
		jwtcookie.WithAudience("wingbox"),
		jwtcookie.WithSameSite(http.SameSiteLaxMode),
		jwtcookie.WithCookieName("__Http-DO_NOT_SHARE-refresh_token"),
	)
}

func NewAuthService() *AuthService {
	s := server.Init()
	clientId := shared.Ensureenv("DISCORD_CLIENT_ID")
	clientSecret := shared.Ensureenv("DISCORD_CLIENT_SECRET")
	redirectURI := shared.Ensureenv("REDIRECT_URI")

	jwtKey := []byte(shared.Ensureenv("JWT_SECRET"))
	jwtSalt := []byte(shared.Ensureenv("JWT_SALT"))

	accessMgr, err := newAccessManager(jwtKey, jwtSalt)
	if err != nil {
		s.LogFatal("could not initialize access token cookie manager", "err", err)
	}

	refreshMgr, err := newRefreshManager(jwtKey, jwtSalt)
	if err != nil {
		s.LogFatal("could not initialize refresh token cookie manager", "err", err)
	}

	return &AuthService{s, redirectURI, clientId, clientSecret, accessMgr, refreshMgr}
}

func (as *AuthService) RegisterRoutes() {
	as.server.Handle("/discord", as.server.BaseChain.ThenFunc(as.Discord))
	as.server.Handle("/redirect", as.server.BaseChain.ThenFunc(as.Redirect))
}

func (as *AuthService) Listen(port uint64) {
	as.server.Listen(port)
}