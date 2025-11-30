package erase_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/erase"
	mocks "github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/erase/mocks"
	"github.com/n0f4ph4mst3r/goshort/internal/sl/sldiscard"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

func TestEraseHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		mockUrl      string
		expectedCode int
		mockError    error
	}{
		{
			name:         "Success",
			alias:        "some_alias",
			mockUrl:      "https://duckduckgo.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty alias",
			alias:        "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "URL Not Found",
			alias:        "some_alias",
			expectedCode: http.StatusNotFound,
			mockError:    storage.ErrUrlNotFound,
		},
		{
			name:         "DeleteURL Error",
			alias:        "some_alias",
			expectedCode: http.StatusInternalServerError,
			mockError:    errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlEraserMock := mocks.NewMockUrlEraser(t)

			if tc.expectedCode == http.StatusBadRequest {
				req, _ := http.NewRequest(http.MethodDelete, "/", nil)
				rr := httptest.NewRecorder()

				handler := erase.New(sldiscard.NewDiscardLogger(), urlEraserMock)
				handler.ServeHTTP(rr, req)

				require.Equal(t, tc.expectedCode, rr.Code)
				return
			}

			urlEraserMock.On("DeleteURL", mock.Anything, tc.alias).
				Return(tc.mockUrl, tc.mockError).Once()

			router := chi.NewRouter()
			router.Delete("/url/{alias}", erase.New(sldiscard.NewDiscardLogger(), urlEraserMock))

			req, err := http.NewRequest(http.MethodDelete, "/url/"+tc.alias, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedCode == http.StatusOK {
				var resp erase.Response
				body := rr.Body.String()
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				require.Equal(t, tc.mockUrl, resp.URL)
			}
		})
	}
}
