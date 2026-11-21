package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/la0rg/lingo/internal/config"
)

// wordLimiter limits how many words a client can request per minute on the public endpoints.
// Every word may cost an OpenAI call, so the limit counts words rather than requests.
type wordLimiter struct {
	rl *httprate.RateLimiter
}

func newWordLimiter(wordsPerMinute int) *wordLimiter {
	return &wordLimiter{rl: httprate.NewRateLimiter(wordsPerMinute, time.Minute)}
}

// Allow reports whether the client may request n more words and, if so, counts them.
// It relies on one of the chi ClientIPFrom* middlewares to resolve the client IP.
func (l *wordLimiter) Allow(w http.ResponseWriter, r *http.Request, n int) bool {
	key := httprate.CanonicalizeIP(middleware.GetClientIP(r.Context()))
	r = r.WithContext(httprate.WithIncrement(r.Context(), n))
	return !l.rl.OnLimit(w, r, key)
}

// clientIP resolves the client IP from a trusted proxy header when configured, from RemoteAddr otherwise.
func clientIP(limits config.Limits) func(http.Handler) http.Handler {
	if limits.ClientIPHeader == "" {
		return middleware.ClientIPFromRemoteAddr
	}
	return middleware.ClientIPFromHeader(limits.ClientIPHeader, limits.TrustedProxies...)
}
