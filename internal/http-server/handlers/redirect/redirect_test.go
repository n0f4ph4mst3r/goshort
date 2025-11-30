package redirect_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/redirect"
	mocks "github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/redirect/mocks"
	"github.com/n0f4ph4mst3r/goshort/internal/sl/sldiscard"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		expectedCode int
		mockUrl      string
		mockError    error
	}{
		{
			name:         "Success",
			alias:        "GoDuck",
			expectedCode: http.StatusFound,
			mockUrl:      "https://duckduckgo.com",
		},
		{
			name:         "Empty alias",
			alias:        "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "URL not found",
			alias:        "some_alias",
			expectedCode: http.StatusNotFound,
			mockError:    storage.ErrUrlNotFound,
		},
		{
			name:         "GetURL Error",
			alias:        "some_alias",
			expectedCode: http.StatusInternalServerError,
			mockError:    errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewMockUrlGetter(t)

			if tc.expectedCode == http.StatusBadRequest {
				req, _ := http.NewRequest(http.MethodGet, "/", nil)
				rr := httptest.NewRecorder()

				handler := redirect.New(sldiscard.NewDiscardLogger(), urlGetterMock)
				handler.ServeHTTP(rr, req)

				require.Equal(t, tc.expectedCode, rr.Code)
				return
			}

			urlGetterMock.On("GetURL", mock.Anything, tc.alias).
				Return(tc.mockUrl, tc.mockError).Once()

			router := chi.NewRouter()
			router.Get("/url/{alias}", redirect.New(sldiscard.NewDiscardLogger(), urlGetterMock))

			req, err := http.NewRequest(http.MethodGet, "/url/"+tc.alias, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedCode == http.StatusFound {
				require.Equal(t, tc.mockUrl, rr.Header().Get("Location"))
			}
		})
	}
}
