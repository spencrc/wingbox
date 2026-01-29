package middleware

import (
	"log/slog"
	"net/http"
)

func LogRequest(logger *slog.Logger) func(http.Handler) http.Handler {
	// we need to return a func like this so it can be chained nicely
	//  effectively, pass logger function -> return handler function -> do the middleware!
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request received", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}