package save

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/n0f4ph4mst3r/goshort/internal/http-server/response"
	"github.com/n0f4ph4mst3r/goshort/internal/sl"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Message
	URL   string `json:"url,omitempty"`
	Alias string `json:"alias,omitempty"`
}

type UrlSaver interface {
	SaveURL(ctx context.Context, originalURL, alias string) error
}

type AliasGenerator interface {
	Generate() string
}

func New(log *slog.Logger, saver UrlSaver, gen AliasGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := sl.Init(log, r.Context(), "http-server.handlers.url.save.New")

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			sl.WriteResponse(log, w, r, http.StatusBadRequest,
				response.Error("empty request"),
				"request body is empty")

			return
		}
		if err != nil {
			sl.WriteResponse(log, w, r, http.StatusBadRequest,
				response.Error("invalid request body"),
				"failed to decode request body", sl.Err(err))

			return
		}

		log.Info("request body decoded successfully", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			sl.WriteResponse(log, w, r, http.StatusBadRequest,
				response.ValidationError(validateErr),
				"request validation failed", sl.Err(validateErr))

			return
		}

		alias := req.Alias
		if alias != "" {
			err = saver.SaveURL(r.Context(), req.URL, alias)
			if errors.Is(err, storage.ErrUrlExists) {
				sl.WriteResponse(log, w, r, http.StatusConflict,
					response.Error("alias already exists"),
					"failed to save URL", slog.String("url", req.URL))

				return
			}
			if err != nil {
				sl.WriteResponse(log, w, r, http.StatusInternalServerError,
					response.Error("internal server error"),
					"failed to save URL", sl.Err(err))

				return
			}
		} else {
			if gen == nil {
				gen = &DefaultRandomAlias{}
			}

			for attempt := 0; attempt < 10; attempt++ {
				alias = gen.Generate()

				err = saver.SaveURL(r.Context(), req.URL, alias)
				if err == nil {
					break
				}
				if errors.Is(err, storage.ErrUrlExists) {
					log.Warn("alias collision, regenerating", slog.String("alias", alias), slog.Int("attempt", attempt+1))
					continue
				}

				sl.WriteResponse(log, w, r, http.StatusInternalServerError,
					response.Error("internal server error"),
					"failed to save URL", sl.Err(err))

				return
			}

			if err != nil {
				sl.WriteResponse(log, w, r, http.StatusInternalServerError,
					response.Error("alias collision after multiple attempts, try again later"),
					"failed to save URL", sl.Err(err))

				return
			}
		}

		sl.WriteResponse(log, w, r, 0, Response{
			Message: response.OK(),
			URL:     req.URL,
			Alias:   alias,
		}, "URL saved successfully", slog.String("url", req.URL), slog.String("alias", alias))
	}
}

type DefaultRandomAlias struct{}

func (g *DefaultRandomAlias) Generate() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, 6)
	for i := range buf {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		buf[i] = chars[num.Int64()]
	}
	return string(buf)
}
