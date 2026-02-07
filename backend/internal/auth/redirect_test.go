package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedeemCookie(t *testing.T) {
	var tests = []struct{
		name string
		cookieValue string
		queryState string
		queryCode string
		expectedError error
	}{
		{
			name: "no errors",
			cookieValue:  "secret_state",
			queryState:   "secret_state",
			queryCode:    "auth_code_123",
			expectedError: nil,
		},
		{
			name: "mismatched states",
			cookieValue:  "secret_state",
			queryState:   "wrong_state",
			queryCode:    "auth_code_123",
			expectedError: ErrInvalidState,
		},
		{
			name: "missing cookie",
			cookieValue:  "",
			queryState:   "secret_state",
			queryCode:    "auth_code_123",
			expectedError: http.ErrNoCookie,
		},
		{
			name: "missing code",
			cookieValue:  "secret_state",
			queryState:   "secret_state",
			queryCode:    "",
			expectedError: ErrMissingCode,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T) {
			url := fmt.Sprintf("/redirect?state=%s&code=%s", test.queryState, test.queryCode)
			req := httptest.NewRequest("GET", url, nil)

			if test.cookieValue != "" {
				cookie := generateStateCookie(test.cookieValue)
				req.AddCookie(&cookie)
			}

			code, err := redeemCodeFromCookie(req)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("got error %v, but was expecting error %v! with cookie value %s, query state %s, and query code %s", err, test.expectedError, test.cookieValue, test.queryState, test.queryCode)
			} else if code != test.queryCode && err == nil {
				t.Errorf("got code %s, but was expecting code %s! with cookie value %s, query state %s", code, test.queryCode, test.cookieValue, test.queryState)
			}
		})
	}
}