package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRF(t *testing.T) {
	t.Parallel()

	handler := CSRF()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name         string
		method       string
		secFetchSite string
		origin       string
		want         int
	}{
		{"same-origin POST allowed", http.MethodPost, "same-origin", "", http.StatusOK},
		{"cross-site POST blocked", http.MethodPost, "cross-site", "", http.StatusForbidden},
		{"mismatched origin POST blocked", http.MethodPost, "", "https://attacker.example", http.StatusForbidden},
		// Non-browser clients send neither header; Origin-less requests are allowed.
		{"headerless POST allowed", http.MethodPost, "", "", http.StatusOK},
		{"cross-site GET allowed", http.MethodGet, "cross-site", "", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "https://bitofbytes.io/contact", nil)
			if tc.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tc.secFetchSite)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Errorf("got status %d, want %d", rr.Code, tc.want)
			}
		})
	}
}
