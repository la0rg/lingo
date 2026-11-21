package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	limiterIdleTTL       = 10 * time.Minute
	limiterSweepInterval = time.Minute
)

// WordLimiter rate-limits the number of words a single client can request.
// Every word may cost an OpenAI call, so the limit is in words rather than in requests.
type WordLimiter struct {
	mu        sync.Mutex
	clients   map[string]*clientLimiter
	limit     rate.Limit
	burst     int
	ipHeader  string
	now       func() time.Time
	lastSweep time.Time
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewWordLimiter creates a limiter that allows wordsPerMinute words per client, with the given burst.
// ipHeader is an optional trusted proxy header holding the client IP.
func NewWordLimiter(wordsPerMinute, burst int, ipHeader string) *WordLimiter {
	return &WordLimiter{
		clients:  make(map[string]*clientLimiter),
		limit:    rate.Limit(float64(wordsPerMinute) / 60),
		burst:    burst,
		ipHeader: ipHeader,
		now:      time.Now,
	}
}

// Allow reports whether the client making r may request n more words.
func (l *WordLimiter) Allow(r *http.Request, n int) bool {
	key := l.clientIP(r)
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(now)

	c, ok := l.clients[key]
	if !ok {
		c = &clientLimiter{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.clients[key] = c
	}
	c.lastSeen = now

	return c.limiter.AllowN(now, n)
}

// sweep drops limiters of clients that have been idle for a while. It must be called with l.mu held.
func (l *WordLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < limiterSweepInterval {
		return
	}
	l.lastSweep = now

	for key, c := range l.clients {
		if now.Sub(c.lastSeen) > limiterIdleTTL {
			delete(l.clients, key)
		}
	}
}

func (l *WordLimiter) clientIP(r *http.Request) string {
	if l.ipHeader != "" {
		if ip := strings.TrimSpace(r.Header.Get(l.ipHeader)); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
