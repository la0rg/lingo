package handler

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/config"
)

// limitedRequest runs a request through the client IP middleware and the word limiter.
func limitedRequest(limits config.Limits, l *wordLimiter, remoteAddr, ipHeaderValue string, words int) bool {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	if ipHeaderValue != "" {
		req.Header.Set("CF-Connecting-IP", ipHeaderValue)
	}

	var allowed bool
	clientIP(limits)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed = l.Allow(w, r, words)
	})).ServeHTTP(httptest.NewRecorder(), req)

	return allowed
}

func TestWordLimiter_CountsWords(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(10)
		limits := config.Limits{}

		assert.True(t, limitedRequest(limits, l, "192.0.2.1:1", "", 6))
		assert.True(t, limitedRequest(limits, l, "192.0.2.1:2", "", 4), "port must not matter")
		assert.False(t, limitedRequest(limits, l, "192.0.2.1:1", "", 1), "limit exhausted")
		assert.True(t, limitedRequest(limits, l, "192.0.2.2:1", "", 10), "other clients are not affected")
		assert.False(t, limitedRequest(limits, l, "192.0.2.3:1", "", 11), "more than the limit is never allowed")

		time.Sleep(2 * time.Minute)
		assert.True(t, limitedRequest(limits, l, "192.0.2.1:1", "", 10), "limit resets")
	})
}

func TestWordLimiter_IgnoresIPHeaderByDefault(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(1)
		limits := config.Limits{}

		assert.True(t, limitedRequest(limits, l, "192.0.2.1:1", "198.51.100.1", 1))
		assert.False(t, limitedRequest(limits, l, "192.0.2.1:1", "198.51.100.2", 1),
			"spoofed header must not bypass the limit")
	})
}

func TestWordLimiter_TrustedProxy(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(1)
		limits := config.Limits{ClientIPHeader: "CF-Connecting-IP", TrustedProxies: []string{"10.0.0.0/8"}}

		assert.True(t, limitedRequest(limits, l, "10.0.0.1:1", "198.51.100.1", 1))
		assert.True(t, limitedRequest(limits, l, "10.0.0.1:1", "198.51.100.2", 1), "clients behind the proxy differ")
		assert.False(t, limitedRequest(limits, l, "10.0.0.2:1", "198.51.100.1", 1))

		assert.True(t, limitedRequest(limits, l, "192.0.2.9:1", "198.51.100.3", 1))
		assert.False(t, limitedRequest(limits, l, "192.0.2.9:1", "198.51.100.4", 1),
			"header from an untrusted address is ignored")
	})
}

func TestResolveClientIP(t *testing.T) {
	t.Parallel()

	cloudflare := []netip.Prefix{netip.MustParsePrefix("173.245.48.0/20"), netip.MustParsePrefix("2400:cb00::/32")}

	cases := []struct {
		name       string
		remoteAddr string
		header     string
		values     []string
		proxies    []netip.Prefix
		want       string
	}{
		{"remote address", "192.0.2.1:1234", "", nil, nil, "192.0.2.1"},
		{"remote address without port", "192.0.2.1", "", nil, nil, "192.0.2.1"},
		{"ipv6 remote address", "[2001:db8::1]:1234", "", nil, nil, "2001:db8::1"},
		{"ipv4-mapped ipv6 is folded", "[::ffff:192.0.2.1]:1", "", nil, nil, "192.0.2.1"},
		{"invalid remote address", "garbage", "", nil, nil, ""},
		{"header ignored when not configured", "192.0.2.1:1", "", []string{"198.51.100.1"}, nil, "192.0.2.1"},
		{"header from trusted proxy", "173.245.48.5:1", "CF-Connecting-IP", []string{"198.51.100.1"}, cloudflare,
			"198.51.100.1"},
		{"ipv6 header from ipv6 proxy", "[2400:cb00::1]:1", "CF-Connecting-IP", []string{"2001:db8::7"}, cloudflare,
			"2001:db8::7"},
		{"header from untrusted address", "192.0.2.1:1", "CF-Connecting-IP", []string{"198.51.100.1"}, cloudflare,
			"192.0.2.1"},
		{"no prefixes trusts every address", "192.0.2.1:1", "CF-Connecting-IP", []string{"198.51.100.1"}, nil,
			"198.51.100.1"},
		{"missing header falls back to remote", "173.245.48.5:1", "CF-Connecting-IP", nil, cloudflare,
			"173.245.48.5"},
		{"last header value wins", "173.245.48.5:1", "CF-Connecting-IP", []string{"6.6.6.6", " 198.51.100.1 "},
			cloudflare, "198.51.100.1"},
		{"invalid header fails closed", "173.245.48.5:1", "CF-Connecting-IP", []string{"198.51.100.1", "nope"},
			cloudflare, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = c.remoteAddr
			for _, v := range c.values {
				req.Header.Add("CF-Connecting-IP", v)
			}

			assert.Equal(t, c.want, resolveClientIP(req, c.header, c.proxies))
		})
	}
}

func TestWordLimiter_GroupsIPv6By64(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(1)
		limits := config.Limits{}

		assert.True(t, limitedRequest(limits, l, "[2001:db8:1:2::1]:1", "", 1))
		assert.False(t, limitedRequest(limits, l, "[2001:db8:1:2:ffff::9]:1", "", 1), "same /64 shares the limit")
		assert.True(t, limitedRequest(limits, l, "[2001:db8:1:3::1]:1", "", 1), "other /64 is another client")
	})
}

func TestLimiterKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "192.0.2.1", limiterKey("192.0.2.1"))
	assert.Equal(t, "2001:db8:1:2::/64", limiterKey("2001:db8:1:2:abcd::1"))
	assert.Equal(t, "", limiterKey(""))
}

func TestWordLimiter_RetryAfter(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(5) // one word every 12 seconds
		allow := func(n int) (bool, string) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			var ok bool
			clientIP(config.Limits{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ok = l.Allow(w, r, n)
			})).ServeHTTP(rec, req)
			return ok, rec.Header().Get("Retry-After")
		}

		ok, _ := allow(5)
		require.True(t, ok)

		ok, retryAfter := allow(1)
		assert.False(t, ok)
		assert.Equal(t, "12", retryAfter)

		ok, retryAfter = allow(3)
		assert.False(t, ok)
		assert.Equal(t, "36", retryAfter, "rejected requests don't consume words")

		ok, retryAfter = allow(6)
		assert.False(t, ok)
		assert.Equal(t, "60", retryAfter, "more than the limit is never allowed")

		time.Sleep(12 * time.Second)
		ok, _ = allow(1)
		assert.True(t, ok, "words refill continuously")
	})
}

func TestWordLimiter_DropsIdleClients(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		l := newWordLimiter(2)
		limits := config.Limits{}

		limitedRequest(limits, l, "192.0.2.1:1", "", 2)
		limitedRequest(limits, l, "192.0.2.2:1", "", 1)
		assert.Len(t, l.clients, 2)

		time.Sleep(30 * time.Second)
		limitedRequest(limits, l, "192.0.2.3:1", "", 1)
		assert.Len(t, l.clients, 3, "clients are kept until their bucket refills")

		time.Sleep(40 * time.Second)
		limitedRequest(limits, l, "192.0.2.3:1", "", 1)
		assert.Len(t, l.clients, 1)
		assert.Contains(t, l.clients, "192.0.2.3")

		assert.True(t, limitedRequest(limits, l, "192.0.2.1:1", "", 2), "a dropped client starts with a full bucket")
	})
}

func TestWordLimiter_Disabled(t *testing.T) {
	t.Parallel()

	l := newWordLimiter(0)
	assert.True(t, limitedRequest(config.Limits{}, l, "192.0.2.1:1", "", 1000))
	assert.Empty(t, l.clients)
}
