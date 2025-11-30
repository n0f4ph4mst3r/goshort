package erase

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/response"
	"github.com/n0f4ph4mst3r/goshort/internal/sl"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

type Response struct {
	response.Message
	URL string `json:"url,omitempty"`
}

type UrlEraser interface {
	DeleteURL(ctx context.Context, alias string) (string, error)
}

func New(log *slog.Logger, urlEraser UrlEraser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := sl.Init(log, r.Context(), "http-server.handlers.url.erase.New")

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			sl.WriteResponse(log, w, r, http.StatusBadRequest,
				response.Error("invalid request"),
				"alias is empty")

			return
		}

		url, err := urlEraser.DeleteURL(r.Context(), alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			sl.WriteResponse(log, w, r, http.StatusNotFound,
				response.Error("invalid request"),
				"URL not found", slog.String("alias", alias))

			return
		}
		if err != nil {
			sl.WriteResponse(log, w, r, http.StatusInternalServerError,
				response.Error("internal error"),
				"failed to delete URL", sl.Err(err))

			return
		}

		sl.WriteResponse(log, w, r, 0,
			Response{
				Message: response.OK(),
				URL:     url,
			},
			"URL deleted", slog.String("url", url))
	}
}
