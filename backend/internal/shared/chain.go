package shared

import (
	"net/http"
	"slices"
)

type Chain []func(http.Handler) http.Handler

func (c Chain) ThenFunc(handler http.HandlerFunc) http.Handler {
	return c.then(handler)
}

func (c Chain) then(handler http.Handler) http.Handler {
	for _, middleware := range slices.Backward(c) {
		handler = middleware(handler)
	}
	return handler
}