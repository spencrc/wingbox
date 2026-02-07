package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestFetchTokenData(t *testing.T) {
	const ACCESS_TOKEN = "mock_atoken_123"
	const REFRESH_TOKEN = "mock_rtoken_456"
	
	body := fmt.Sprintf(`{
		"access_token": "%s",
		"refresh_token": "%s"
	}`, ACCESS_TOKEN, REFRESH_TOKEN)

	// see here: https://dev.to/andreidascalu/testing-your-api-client-in-go-a-method-4bm4
	client := &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(strings.NewReader(body)),
				Header: make(http.Header),
			}
		}),
	}

	res, err := fetchTokenData("code", "uri", "id", "secret", client)
	if err != nil {
		t.Errorf("did not expect error, got %v", err)
	}

	if res.AccessToken != ACCESS_TOKEN {
		t.Errorf("expected %s, got %s", ACCESS_TOKEN, res.AccessToken)
	}
	if res.RefreshToken != "mock_rtoken_456" {
		t.Errorf("expected %s, got %s", REFRESH_TOKEN, res.RefreshToken)
	}
}