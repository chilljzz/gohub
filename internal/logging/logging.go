package logging

import (
	"context"
	"log/slog"
	"os"
)

type requestIDkey struct{}

func New(
	mode string,
) *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}

	var handler slog.Handler

	if mode == "release" {

		handler = slog.NewJSONHandler(os.Stdout, opts)

	} else {

		opts.Level = slog.LevelDebug

		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(slog.String("service", "gohub"))
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDkey{}, requestID)

}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(
		requestIDkey{},
	).(string)

	return requestID
}

func FromContext(
	ctx context.Context,
) *slog.Logger {

	logger := slog.Default()

	requestID := RequestID(ctx)

	if requestID == "" {
		return logger
	}

	return logger.With(
		slog.String(
			"request_id",
			requestID,
		),
	)
}
