package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/handler/mocks"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/repository"
	"github.com/la0rg/lingo/internal/service"
	"github.com/la0rg/lingo/web"
)

const (
	testAPIKey         = "test-key"
	testWordsPerMinute = 5
)

func testCard(word string) *model.Card {
	return &model.Card{
		Word:         word,
		Translations: model.Translations{{PartOfSpeech: "Noun", Translations: []string{"t-" + word}}},
		Context:      "context of " + word,
	}
}

type testServer struct {
	handler http.Handler
	cards   *mocks.CardService
	apkg    *mocks.APKGService
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	templates, err := web.NewTemplates()
	require.NoError(t, err)

	collections, err := repository.NewCollection()
	require.NoError(t, err)

	conf := &config.Config{
		APIKeys: []string{"other-key", testAPIKey},
		Limits: config.Limits{
			MaxWords:       3,
			MaxWordLength:  10,
			WordsPerMinute: testWordsPerMinute,
		},
	}

	s := &testServer{cards: mocks.NewCardService(t), apkg: mocks.NewAPKGService(t)}
	s.handler = NewRouter(conf, s.cards, s.apkg, templates, collections)
	return s
}

func (s *testServer) do(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	return rec
}

// expectCard makes the card service return a test card for word.
func (s *testServer) expectCard(word string) {
	s.cards.EXPECT().Get(mock.Anything, mock.Anything, word).Return(testCard(word), nil).Once()
}

// expectAPKG expects one CreateOnS3 call with n words and records the collection slug and the words.
func (s *testServer) expectAPKG(
	n int, file *model.CollectionFile, err error,
) (slug *string, words *[]string) {
	slug, words = new(string), new([]string)

	// the generated mock passes variadic words as a single []string argument
	nWords := mock.MatchedBy(func(w []string) bool { return len(w) == n })

	s.apkg.EXPECT().CreateOnS3(mock.Anything, mock.Anything, nWords).
		RunAndReturn(func(_ context.Context, c *model.Collection, w ...string) (*model.CollectionFile, error) {
			*slug, *words = model.CollectionSlug(c.SourceLanguage, c.TargetLanguage), w
			if err != nil {
				return nil, err
			}
			if file != nil {
				return file, nil
			}
			return &model.CollectionFile{URL: "https://s3.example/file.apkg", CardsCount: len(w)}, nil
		}).Once()

	return slug, words
}

func apiRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	return req
}

func formRequest(target string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func collectionForm(words string) url.Values {
	return url.Values{"source": {"de"}, "target": {"en"}, "words": {words}}
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) HTTPError {
	t.Helper()
	var httpErr HTTPError
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&httpErr))
	return httpErr
}

func TestAPIKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		header string
		code   int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + testAPIKey, http.StatusUnauthorized},
		{"malformed", "Bearer", http.StatusUnauthorized},
		{"unknown key", "Bearer nope", http.StatusUnauthorized},
		{"key prefix", "Bearer test", http.StatusUnauthorized},
		{"valid key", "Bearer " + testAPIKey, http.StatusOK},
		{"another valid key", "Bearer other-key", http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)
			if c.code == http.StatusOK {
				s.expectCard("apfel")
			}

			req := httptest.NewRequest(http.MethodGet, "/api/collections/de-en/cards/apfel", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}

			assert.Equal(t, c.code, s.do(req).Code)
		})
	}
}

func TestValidAPIKey(t *testing.T) {
	t.Parallel()

	assert.False(t, validAPIKey(nil, ""))
	assert.False(t, validAPIKey([]string{"a"}, ""))
	assert.False(t, validAPIKey([]string{"abc"}, "ab"))
	assert.True(t, validAPIKey([]string{"x", "abc"}, "abc"))
}

func TestCardAPI(t *testing.T) {
	t.Parallel()

	t.Run("returns the card", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectCard("apfel")

		rec := s.do(apiRequest(http.MethodGet, "/api/collections/DE-EN/cards/%20apfel%20", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		var card model.Card
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&card))
		assert.Equal(t, "apfel", card.Word)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.cards.EXPECT().Get(mock.Anything, mock.Anything, "apfel").Return(nil, errors.New("db down"))

		rec := s.do(apiRequest(http.MethodGet, "/api/collections/de-en/cards/apfel", nil))

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, ErrInternal.Message, decodeError(t, rec).Message)
	})

	cases := []struct {
		name    string
		target  string
		code    int
		message string
	}{
		{"unknown collection", "/api/collections/xx-yy/cards/apfel", 404, ErrCollectionNotFound.Message},
		{"blank word", "/api/collections/de-en/cards/%20", 400, ErrEmptyWord.Message},
		{"word too long", "/api/collections/de-en/cards/abcdefghijk", 400, ErrWordTooLong(10).Message},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t) // no expectations: the card service must not be called

			rec := s.do(apiRequest(http.MethodGet, c.target, nil))

			require.Equal(t, c.code, rec.Code)
			assert.Equal(t, c.message, decodeError(t, rec).Message)
		})
	}
}

func TestAPKGAPI(t *testing.T) {
	t.Parallel()

	t.Run("creates the file with normalized words", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		slug, words := s.expectAPKG(2, nil, nil)

		body := `{"words": ["Apfel", " apfel ", "", "Haus"]}`
		rec := s.do(apiRequest(http.MethodPost, "/api/collections/de-en/apkg", strings.NewReader(body)))

		require.Equal(t, http.StatusOK, rec.Code)
		var resp model.APKGResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "https://s3.example/file.apkg", resp.URL)
		assert.Equal(t, "de-en", *slug)
		assert.ElementsMatch(t, []string{"apfel", "haus"}, *words)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectAPKG(1, nil, errors.New("s3 down"))

		rec := s.do(apiRequest(http.MethodPost, "/api/collections/de-en/apkg", strings.NewReader(`{"words": ["a"]}`)))

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, ErrInternal.Message, decodeError(t, rec).Message)
	})

	cases := []struct {
		name    string
		target  string
		body    string
		code    int
		message string
	}{
		{"unknown collection", "/api/collections/xx-yy/apkg", `{"words": ["a"]}`, 404, ErrCollectionNotFound.Message},
		{"invalid json", "", `{"words":`, 400, "Invalid request format"},
		{"no words", "", `{"words": []}`, 400, ErrEmptyWordList.Message},
		{"blank words", "", `{"words": [" ", ""]}`, 400, ErrEmptyWordList.Message},
		{"too many words", "", `{"words": ["a", "b", "c", "d"]}`, 400, ErrTooManyWords(3).Message},
		{"word too long", "", `{"words": ["abcdefghijk"]}`, 400, ErrWordTooLong(10).Message},
		{"body too large", "", `{"words": ["` + strings.Repeat("a", maxBodyBytes) + `"]}`, 413,
			ErrBodyTooLarge.Message},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			target := c.target
			if target == "" {
				target = "/api/collections/de-en/apkg"
			}
			rec := s.do(apiRequest(http.MethodPost, target, strings.NewReader(c.body)))

			require.Equal(t, c.code, rec.Code)
			assert.Equal(t, c.message, decodeError(t, rec).Message)
		})
	}
}

func TestCardPreview(t *testing.T) {
	t.Parallel()

	t.Run("renders the card", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectCard("Apfel")

		rec := s.do(httptest.NewRequest(http.MethodGet, "/card?source=de&target=en&word=Apfel", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "t-Apfel")
	})

	t.Run("escapes card content", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.cards.EXPECT().Get(mock.Anything, mock.Anything, "x").Return(testCard("<script>x</script>"), nil)

		rec := s.do(httptest.NewRequest(http.MethodGet, "/card?source=en&target=de&word=x", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "<script>x</script>")
		assert.Contains(t, rec.Body.String(), "&lt;script&gt;")
	})

	t.Run("card not created", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.cards.EXPECT().Get(mock.Anything, mock.Anything, "qwrt").Return(nil, nil)

		rec := s.do(httptest.NewRequest(http.MethodGet, "/card?source=de&target=en&word=qwrt", nil))

		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrInvalidWord.Message)
	})

	cases := []struct {
		name    string
		query   string
		code    int
		message string
	}{
		{"missing languages", "word=apfel", 400, ErrInvalidLanguage.Message},
		{"unknown collection", "source=xx&target=en&word=apfel", 404, ErrCollectionNotFound.Message},
		{"missing word", "source=de&target=en", 400, ErrEmptyWord.Message},
		{"word too long", "source=de&target=en&word=abcdefghijk", 400, ErrWordTooLong(10).Message},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			rec := s.do(httptest.NewRequest(http.MethodGet, "/card?"+c.query, nil))

			require.Equal(t, c.code, rec.Code)
			assert.Contains(t, rec.Body.String(), c.message)
		})
	}
}

func TestCardPreview_RateLimit(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		s := newTestServer(t)
		preview := func(word string) *httptest.ResponseRecorder {
			return s.do(httptest.NewRequest(http.MethodGet, "/card?source=de&target=en&word="+word, nil))
		}

		for i := range testWordsPerMinute {
			word := fmt.Sprintf("w%d", i)
			s.expectCard(word)
			require.Equal(t, http.StatusOK, preview(word).Code, "request %d", i)
		}

		// no expectation: a rate limited request must not reach the card service
		rec := preview("over")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrRateLimited.Message)
		assert.NotEmpty(t, rec.Header().Get("Retry-After"))

		other := httptest.NewRequest(http.MethodGet, "/card?source=de&target=en&word=other", nil)
		other.RemoteAddr = "198.51.100.7:1234"
		s.expectCard("other")
		assert.Equal(t, http.StatusOK, s.do(other).Code, "other clients are not affected")

		time.Sleep(2 * time.Minute)
		s.expectCard("later")
		assert.Equal(t, http.StatusOK, preview("later").Code, "limit resets after the window")
	})
}

func TestCreateCollection(t *testing.T) {
	t.Parallel()

	t.Run("creates the collection", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		file := &model.CollectionFile{URL: "https://s3.example/c.apkg", CardsCount: 1, MissingWords: []string{"qwrt"}}
		slug, words := s.expectAPKG(2, file, nil)

		rec := s.do(formRequest("/collection", collectionForm("Apfel, qwrt\n apfel \n")))

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "https://s3.example/c.apkg")
		assert.Contains(t, body, "qwrt")
		assert.Contains(t, body, "1/2")
		assert.Equal(t, "de-en", *slug)
		assert.ElementsMatch(t, []string{"apfel", "qwrt"}, *words)
	})

	t.Run("no valid words", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectAPKG(1, nil, fmt.Errorf("wrapped: %w", service.ErrEmptyCards))

		rec := s.do(formRequest("/collection", collectionForm("qwrt")))

		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrNoValidWords.Message)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectAPKG(1, nil, errors.New("boom"))

		rec := s.do(formRequest("/collection", collectionForm("a")))

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrInternal.Message)
	})

	cases := []struct {
		name    string
		form    url.Values
		code    int
		message string
	}{
		{"missing languages", url.Values{"words": {"a"}}, 400, ErrInvalidLanguage.Message},
		{"unknown collection", url.Values{"source": {"de"}, "target": {"xx"}, "words": {"a"}}, 404,
			ErrCollectionNotFound.Message},
		{"no words", collectionForm(""), 400, ErrEmptyWordList.Message},
		{"blank words", collectionForm(" ,\n, "), 400, ErrEmptyWordList.Message},
		{"too many words", collectionForm("a,b,c,d"), 400, ErrTooManyWords(3).Message},
		{"word too long", collectionForm("abcdefghijk"), 400, ErrWordTooLong(10).Message},
		{"body too large", collectionForm(strings.Repeat("a,", maxBodyBytes)), 413, ErrBodyTooLarge.Message},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			rec := s.do(formRequest("/collection", c.form))

			require.Equal(t, c.code, rec.Code)
			assert.Contains(t, rec.Body.String(), c.message)
		})
	}
}

func TestCreateCollection_RateLimitCountsWords(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		s := newTestServer(t)

		s.expectAPKG(3, nil, nil)
		require.Equal(t, http.StatusOK, s.do(formRequest("/collection", collectionForm("a,b,c"))).Code)

		// 3 + 3 words exceed the limit of 5 words per minute
		assert.Equal(t, http.StatusTooManyRequests, s.do(formRequest("/collection", collectionForm("d,e,f"))).Code)

		s.expectAPKG(2, nil, nil)
		assert.Equal(t, http.StatusOK, s.do(formRequest("/collection", collectionForm("g,h"))).Code)
	})
}

func TestCollectionPage(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)

	rec := s.do(httptest.NewRequest(http.MethodGet, "/collection", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Create a collection")
}

func TestIndex(t *testing.T) {
	t.Parallel()

	t.Run("default language pair", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectCard("Apfel")

		rec := s.do(httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "t-Apfel")
	})

	t.Run("language from cookie", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.expectCard("Book")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: languageCookie, Value: "en,de"})

		require.Equal(t, http.StatusOK, s.do(req).Code)
	})

	t.Run("unsupported pair resets the cookie", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: languageCookie, Value: "xx,yy"})
		rec := s.do(req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Set-Cookie"), "Max-Age=0")
	})

	t.Run("card service error", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		s.cards.EXPECT().Get(mock.Anything, mock.Anything, "Apfel").Return(nil, errors.New("boom"))

		rec := s.do(httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrInternal.Message)
	})
}

func TestLanguageSelectorFromCookie(t *testing.T) {
	t.Parallel()

	cases := []struct {
		cookie         string
		source, target string
	}{
		{"", "de", "en"},
		{"ru,de", "ru", "de"},
		{"ru", "de", "en"},
		{"ru,", "de", "en"},
		{",de", "de", "en"},
		{"ru,de,en", "de", "en"},
	}

	h := &handlers{}
	for _, c := range cases {
		t.Run(c.cookie, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.cookie != "" {
				req.AddCookie(&http.Cookie{Name: languageCookie, Value: c.cookie})
			}

			ls := h.languageSelectorFromCookie(req)
			assert.Equal(t, c.source, ls.SourceCountryCode)
			assert.Equal(t, c.target, ls.TargetCountryCode)
		})
	}
}

func TestLanguage(t *testing.T) {
	t.Parallel()

	t.Run("sets the cookie", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		req := formRequest("/language", url.Values{"source": {"de"}, "target": {"en"}, "targetNew": {"ru"}})
		req.Header.Set("Referer", "http://localhost/collection")
		rec := s.do(req)

		require.Equal(t, http.StatusOK, rec.Code)
		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, languageCookie, cookies[0].Name)
		assert.Equal(t, "de,ru", cookies[0].Value)
		assert.True(t, cookies[0].Secure)
		assert.True(t, cookies[0].HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
		assert.Empty(t, rec.Header().Get("HX-Refresh"))
	})

	t.Run("refreshes the index page", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		req := formRequest("/language", url.Values{"source": {"de"}, "target": {"en"}, "sourceNew": {"en"}})
		req.Header.Set("Referer", "http://localhost/")
		rec := s.do(req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
		assert.Equal(t, "en,de", rec.Result().Cookies()[0].Value, "same language swaps the pair")
	})

	t.Run("missing languages", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		rec := s.do(formRequest("/language", url.Values{"source": {"de"}}))

		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), ErrInvalidLanguage.Message)
	})
}

func TestStaticRoutes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		target      string
		contentType string
	}{
		{"/styles.css", "text/css; charset=utf-8"},
		{"/public/favicon.svg", "image/svg+xml"},
		{"/public/js/copy_apkg.js", "text/javascript; charset=utf-8"},
		{"/version", "application/json"},
	}

	for _, c := range cases {
		t.Run(c.target, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			rec := s.do(httptest.NewRequest(http.MethodGet, c.target, nil))

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, c.contentType, rec.Header().Get("Content-Type"))
			assert.NotEmpty(t, rec.Body.Bytes())
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)

	rec := s.do(httptest.NewRequest(http.MethodGet, "/version", nil))

	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
}

func TestRequestID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"from request header", map[string]string{requestIDHeader: "req-1", cloudflareRayIDHeader: "ray"}, "req-1"},
		{"from cloudflare ray", map[string]string{cloudflareRayIDHeader: "ray-1"}, "ray-1"},
		{"generated", nil, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			got := s.do(req).Header().Get(requestIDHeader)

			if c.want != "" {
				assert.Equal(t, c.want, got)
			} else {
				assert.Len(t, got, 36, "uuid expected")
			}
		})
	}
}

func TestRedactHeaders(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	h.Set("Authorization", "Bearer secret")
	h.Set("Cookie", "language=de,en")
	h.Set("Accept", "text/html")

	redacted := redactHeaders(h)

	assert.Equal(t, "[REDACTED]", redacted.Get("Authorization"))
	assert.Equal(t, "[REDACTED]", redacted.Get("Cookie"))
	assert.Equal(t, "text/html", redacted.Get("Accept"))
	assert.Empty(t, redacted.Values("Set-Cookie"), "absent headers must not be added")
	assert.Equal(t, "Bearer secret", h.Get("Authorization"), "original headers must not be modified")
}

func TestRecoverer(t *testing.T) {
	t.Parallel()

	h := Recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	rec := httptest.NewRecorder()

	require.NotPanics(t, func() { h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil)) })
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMapError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.Equal(t, ErrEmptyWord, mapError(ctx, ErrEmptyWord))
	assert.Equal(t, ErrEmptyWord, mapError(ctx, fmt.Errorf("wrapped: %w", ErrEmptyWord)))
	assert.Equal(t, ErrBodyTooLarge, mapError(ctx, &http.MaxBytesError{Limit: 1}))
	assert.Equal(t, ErrInternal, mapError(ctx, errors.New("boom")))
}

func TestRequestError(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ErrEmptyWordList, requestError(model.ErrEmptyWordList))
	assert.Equal(t, ErrTooManyWords(7), requestError(&model.TooManyWordsError{Max: 7}))
	assert.Equal(t, ErrWordTooLong(9), requestError(fmt.Errorf("w: %w", &model.WordTooLongError{Max: 9})))
	assert.Equal(t, ErrBodyTooLarge, requestError(&http.MaxBytesError{Limit: 1}))
	assert.Equal(t, BadRequest("Invalid request format"), requestError(errors.New("parse error")))
}

func TestHTTPError(t *testing.T) {
	t.Parallel()

	assert.True(t, BadRequest("x").Is4xx())
	assert.True(t, ErrRateLimited.Is4xx())
	assert.False(t, ErrInternal.Is4xx())
	assert.Equal(t, "Bad Request: x", BadRequest("x").Error())
}
