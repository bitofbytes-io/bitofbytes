package middleware

import "net/http"

// CSRF returns middleware that rejects cross-origin state-changing requests.
//
// It uses the standard library's http.CrossOriginProtection, which checks the
// browser-set Sec-Fetch-Site and Origin headers instead of tokens.
func CSRF() func(http.Handler) http.Handler {
	return http.NewCrossOriginProtection().Handler
}
