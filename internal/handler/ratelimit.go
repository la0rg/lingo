package handler

import (
	"context"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/la0rg/lingo/internal/config"
)

// wordLimiter limits how many words a client can request per minute on the public endpoints.
// Every word may cost an OpenAI call, so the limit counts words rather than requests.
// Each client gets a token bucket that holds up to a minute's worth of words and refills continuously.
type wordLimiter struct {
	limit rate.Limit
	burst int

	mu        sync.Mutex
	clients   map[string]*clientLimiter
	lastSweep time.Time
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// idleTTL is the time an unused bucket needs to refill completely. A full bucket is the same as a new one,
// so it can be dropped without changing the outcome of later requests.
const idleTTL = time.Minute

func newWordLimiter(wordsPerMinute int) *wordLimiter {
	limit := rate.Inf
	if wordsPerMinute > 0 {
		limit = rate.Limit(float64(wordsPerMinute) / idleTTL.Seconds())
	}

	return &wordLimiter{
		limit:   limit,
		burst:   wordsPerMinute,
		clients: make(map[string]*clientLimiter),
	}
}

// Allow reports whether the client may request n more words and, if so, counts them.
// When it returns false, it sets the Retry-After header.
// The client IP is resolved by the clientIP middleware.
func (l *wordLimiter) Allow(w http.ResponseWriter, r *http.Request, n int) bool {
	if l.limit == rate.Inf {
		return true
	}

	now := time.Now()
	limiter := l.client(limiterKey(clientIPFromContext(r.Context())), now)

	reservation := limiter.ReserveN(now, n)
	if !reservation.OK() { // more words than the bucket can ever hold
		w.Header().Set("Retry-After", strconv.Itoa(int(idleTTL.Seconds())))
		return false
	}

	if delay := reservation.DelayFrom(now); delay > 0 {
		reservation.CancelAt(now) // the request is rejected, give the tokens back
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(delay.Seconds()))))
		return false
	}

	return true
}

// client returns the client's bucket, creating it if needed, and drops buckets of idle clients.
func (l *wordLimiter) client(key string, now time.Time) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) >= idleTTL {
		for k, c := range l.clients {
			if now.Sub(c.lastSeen) >= idleTTL {
				delete(l.clients, k)
			}
		}
		l.lastSweep = now
	}

	c, ok := l.clients[key]
	if !ok {
		c = &clientLimiter{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.clients[key] = c
	}
	c.lastSeen = now

	return c.limiter
}

// limiterKey groups IPv6 clients by /64, the usual size of a single customer's allocation,
// so a client can't bypass the limit by rotating addresses within it.
func limiterKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is6() {
		return ip
	}
	return netip.PrefixFrom(addr, 64).Masked().String()
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
