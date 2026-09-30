package middleware

import (
	"net/http"

	csrf "filippo.io/csrf/gorilla"
)

// CSRF returns middleware that rejects cross-origin state-changing requests.
//
// It uses filippo.io/csrf/gorilla, a drop-in replacement for gorilla/csrf that
// checks Sec-Fetch-Site and Origin headers instead of tokens (GHSA-82ff-hg59-8x73).
func CSRF() func(http.Handler) http.Handler {
	return csrf.Protect(nil)
}
