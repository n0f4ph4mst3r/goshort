package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/save"
	mocks "github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/save/mocks"
	"github.com/n0f4ph4mst3r/goshort/internal/sl/sldiscard"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

const urlStr = "https://duckduckgo.com"

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		url          string
		expectedCode int
		respError    string
		mockError    error
	}{
		{
			name:         "Success",
			alias:        "GoDuck",
			url:          urlStr,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty URL",
			alias:        "some_alias",
			url:          "",
			expectedCode: http.StatusBadRequest,
			respError:    "field URL is a required field",
		},
		{
			name:         "Invalid URL",
			url:          "ht!tp://invalid-url",
			expectedCode: http.StatusBadRequest,
			respError:    "field URL is not a valid URL",
		},
		{
			name:         "Alias already exists",
			alias:        "existing_alias",
			url:          urlStr,
			expectedCode: http.StatusConflict,
			respError:    "alias already exists",
			mockError:    storage.ErrUrlExists,
		},
		{
			name:         "Empty alias, first attempt succeeds",
			alias:        "",
			url:          urlStr,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty alias, generation succeeds after 6 collisions",
			alias:        "",
			url:          urlStr,
			expectedCode: http.StatusOK,
		},

		{
			name:         "Empty alias, generation fails after multiple attempts",
			alias:        "",
			url:          urlStr,
			expectedCode: http.StatusInternalServerError,
			respError:    "alias collision after multiple attempts, try again later",
			mockError:    storage.ErrUrlExists,
		},
		{
			name:         "SaveURL Error",
			alias:        "some_alias",
			url:          urlStr,
			expectedCode: http.StatusInternalServerError,
			respError:    "internal server error",
			mockError:    errors.New("unexpected error"),
		},
		{
			name:         "Body is empty",
			alias:        "",
			url:          "",
			expectedCode: http.StatusBadRequest,
			respError:    "field URL is a required field",
		},
		{
			name:         "JSON decode error",
			expectedCode: http.StatusBadRequest,
			respError:    "invalid request body",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewMockUrlSaver(t)
			aliasGenMock := mocks.NewMockAliasGenerator(t)

			if tc.respError == "" || tc.mockError != nil {
				if tc.alias == "" {
					switch tc.name {
					case "Empty alias, generation succeeds after 6 collisions":
						aliasGenMock.On("Generate").Return("collision").Times(6)
						aliasGenMock.On("Generate").Return("unique_alias").Once()

						urlSaverMock.On("SaveURL", mock.Anything, tc.url, "collision").
							Return(storage.ErrUrlExists).Times(6)
						urlSaverMock.On("SaveURL", mock.Anything, tc.url, "unique_alias").
							Return(nil).Once()
					case "Empty alias, generation fails after multiple attempts":
						aliasGenMock.On("Generate").Return("collision").Times(10)
						urlSaverMock.On("SaveURL", mock.Anything, tc.url, "collision").
							Return(tc.mockError).Times(10)
					default:
						aliasGenMock.On("Generate").Return("random_alias").Once()
						urlSaverMock.On("SaveURL", mock.Anything, tc.url, "random_alias").
							Return(nil).Once()
					}
				} else {
					urlSaverMock.On("SaveURL", mock.Anything, tc.url, tc.alias).
						Return(tc.mockError).Once()
				}
			}

			handler := save.New(sldiscard.NewDiscardLogger(), urlSaverMock, aliasGenMock)

			var input string
			if tc.respError == "invalid request body" {
				input = `{"url": "missing_quote, "alias": "malformed"}`
			} else {
				input = fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			}

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			var resp save.Response
			body := rr.Body.String()

			require.Equal(t, rr.Code, tc.expectedCode)
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
