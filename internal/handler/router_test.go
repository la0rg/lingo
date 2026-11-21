package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_Routing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		target string
		code   int
	}{
		{"unknown path is not the index page", http.MethodGet, "/nope", http.StatusNotFound},
		{"unknown api path", http.MethodGet, "/api/collections/de-en/nope", http.StatusNotFound},
		{"wrong method", http.MethodPost, "/version", http.StatusMethodNotAllowed},
		{"wrong api method", http.MethodGet, "/api/collections/de-en/apkg", http.StatusMethodNotAllowed},
		{"missing public file", http.MethodGet, "/public/missing.js", http.StatusNotFound},
		{"head is served for get routes", http.MethodHead, "/version", http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t) // no expectations: no service is called

			rec := s.do(apiRequest(c.method, c.target, nil))

			assert.Equal(t, c.code, rec.Code)
		})
	}
}

func TestRouter_MethodNotAllowedListsAllowedMethods(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)

	rec := s.do(httptest.NewRequest(http.MethodDelete, "/collection", nil))

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Contains(t, rec.Header().Get("Allow"), "GET")
	assert.Contains(t, rec.Header().Get("Allow"), "POST")
}

func TestRouter_PathValuesAreUnescaped(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)
	s.expectCard("an/aus")

	rec := s.do(apiRequest(http.MethodGet, "/api/collections/de-en/cards/an%2Faus", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStatusRecorder(t *testing.T) {
	t.Parallel()

	t.Run("implicit 200", func(t *testing.T) {
		t.Parallel()
		rec := &statusRecorder{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}

		_, err := rec.Write([]byte("hello"))

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.status)
		assert.Equal(t, 5, rec.bytes)
	})

	t.Run("first status wins", func(t *testing.T) {
		t.Parallel()
		rec := &statusRecorder{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}

		rec.WriteHeader(http.StatusTeapot)
		rec.WriteHeader(http.StatusInternalServerError)
		_, _ = rec.Write([]byte("ab"))
		_, _ = rec.Write([]byte("c"))

		assert.Equal(t, http.StatusTeapot, rec.status)
		assert.Equal(t, 3, rec.bytes)
	})

	t.Run("unwraps for ResponseController", func(t *testing.T) {
		t.Parallel()
		inner := httptest.NewRecorder()
		rec := &statusRecorder{ResponseWriter: inner, status: http.StatusOK}

		require.NoError(t, http.NewResponseController(rec).Flush())
		assert.True(t, inner.Flushed)
	})
}

func TestRequestLogger_PassesResponseThrough(t *testing.T) {
	t.Parallel()

	h := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "created", rec.Body.String())
}
