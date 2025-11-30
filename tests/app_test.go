package tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/save"
)

const (
	host = "localhost:8080"
)

func TestGoShort_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/api/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: gofakeit.LetterN(10),
		}).
		WithBasicAuth("myuser", "qwerty").
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

//nolint:funlen
func TestGoShort_SaveRedirectDelete(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "ht!tp://invalid-url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
		{
			name:  "Body is empty",
			alias: "",
			url:   "",
			error: "field URL is a required field",
		},
		{
			name:  "JSON decode error",
			error: "invalid request body",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())

			if tc.name == "JSON decode error" {
				e.POST("/api/url").
					WithBytes([]byte(`{"url": "missing_quote, "alias": "malformed"}`)).
					WithBasicAuth("myuser", "qwerty").
					Expect().
					Status(http.StatusBadRequest).
					Body().Contains("invalid request body")

				return
			}

			req := e.POST("/api/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("myuser", "qwerty")

			if tc.error != "" {
				req.Expect().Status(http.StatusBadRequest).
					Body().Contains(tc.error)
				return
			}

			resp_save := req.Expect().Status(http.StatusOK).JSON().Object()

			alias := tc.alias
			if tc.alias != "" {
				resp_save.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp_save.Value("alias").String().NotEmpty()
				alias = resp_save.Value("alias").String().Raw()
			}

			redirectURL := url.URL{
				Scheme: "http",
				Host:   host,
				Path:   "/api/url/" + alias,
			}

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp_redirect, err := client.Get(redirectURL.String())
			require.NoError(t, err)
			defer resp_redirect.Body.Close()

			require.Equal(t, http.StatusFound, resp_redirect.StatusCode,
				"expected redirect StatusFound (302)")

			location := resp_redirect.Header.Get("Location")
			require.NotEmpty(t, location, "redirect should include Location header")

			require.Equal(t, tc.url, location)

			e.DELETE("/api/url/"+alias).
				WithBasicAuth("myuser", "qwerty").
				Expect().Status(http.StatusOK).JSON().Object().
				Value("url").String().IsEqual(tc.url)

			resp_redirect, err = client.Get(redirectURL.String())
			require.NoError(t, err)
			defer resp_redirect.Body.Close()

			require.Equal(t, http.StatusNotFound, resp_redirect.StatusCode)
		})
	}
}
