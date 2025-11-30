package sl

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Init(log *slog.Logger, ctx context.Context, op string) *slog.Logger {
	return log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func WriteResponse(
	log *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	status int,
	v interface{},
	msg string,
	fields ...any,
) {
	if status != 0 {
		w.WriteHeader(status)
	}
	render.JSON(w, r, v)

	if status >= 500 {
		log.Error(msg, fields...)
	} else {
		log.Info(msg, fields...)
	}

}
