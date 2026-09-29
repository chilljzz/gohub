package middleware

import (
	"log/slog"
	"time"

	"github.com/chilljzz/gohub/internal/logging"
	"github.com/gin-gonic/gin"
)

func AccessLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		status := ctx.Writer.Status()

		route := ctx.FullPath()

		if route == "" {
			route = ctx.Request.URL.Path
		}

		logger := logging.FromContext(ctx.Request.Context())

		attrs := []slog.Attr{
			slog.String("method", ctx.Request.Method),

			slog.String("route", route),

			slog.Int("status", status),

			slog.Int64("latency_ms", time.Since(start).Milliseconds()),

			slog.String("client_ip", ctx.ClientIP()),
		}

		if userID, exists := ctx.Get("userID"); exists {
			attrs = append(attrs, slog.Any("user_id", userID))
		}

		level := slog.LevelInfo

		switch {
		case status >= 500:
			level = slog.LevelError

		case status >= 400:
			level = slog.LevelWarn
		}

		logger.LogAttrs(
			ctx.Request.Context(),
			level,
			"http request",
			attrs...,
		)

	}
}
