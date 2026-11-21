package handler

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

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
// The client IP is resolved by the clientIP middleware.
func (l *wordLimiter) Allow(w http.ResponseWriter, r *http.Request, n int) bool {
	key := httprate.CanonicalizeIP(clientIPFromContext(r.Context()))
	r = r.WithContext(httprate.WithIncrement(r.Context(), n))
	return !l.rl.OnLimit(w, r, key)
}

type clientIPKey struct{}

func clientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

// clientIP stores the client IP in the request context. The IP is taken from the TCP remote address, or from
// limits.ClientIPHeader when the request comes from one of limits.TrustedProxies (any address when none are set).
// Limits must be validated: invalid prefixes are ignored.
func clientIP(limits config.Limits) func(http.Handler) http.Handler {
	var proxies []netip.Prefix
	for _, p := range limits.TrustedProxies {
		if prefix, err := netip.ParsePrefix(p); err == nil {
			proxies = append(proxies, prefix)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := resolveClientIP(r, limits.ClientIPHeader, proxies)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientIPKey{}, ip)))
		})
	}
}

func resolveClientIP(r *http.Request, header string, proxies []netip.Prefix) string {
	remote, ok := parseRemoteAddr(r.RemoteAddr)
	if !ok {
		return ""
	}

	if header == "" || !trusted(remote, proxies) {
		return remote.String()
	}

	values := r.Header.Values(header)
	if len(values) == 0 {
		return remote.String()
	}

	// The last value is set by the proxy closest to us. If it doesn't parse, fail closed instead of
	// falling back to less trusted values.
	ip, err := netip.ParseAddr(strings.TrimSpace(values[len(values)-1]))
	if err != nil {
		return ""
	}
	return ip.Unmap().WithZone("").String()
}

func parseRemoteAddr(remoteAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return ip.Unmap().WithZone(""), true
}

func trusted(ip netip.Addr, proxies []netip.Prefix) bool {
	if len(proxies) == 0 {
		return true
	}
	for _, p := range proxies {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
