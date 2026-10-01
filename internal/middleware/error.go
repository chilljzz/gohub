package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/chilljzz/gohub/internal/logging"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				slog.Error(
					"http panic",
					slog.String("methon", ctx.Request.Method),
					slog.String("path", ctx.Request.URL.Path),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)

				if !ctx.Writer.Written() {
					response.Fail(
						ctx,
						response.CodeServerError,
						"internal server error",
					)
				}

				ctx.Abort()
			}
		}()
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors.Last().Err
		if ctx.Writer.Written() {
			slog.Error(
				"error after response written",
				slog.String("methon", ctx.Request.Method),
				slog.String("path", ctx.Request.URL.Path),
				slog.Any("error", err),
			)

			return
		}

		appErr := mapError(err)

		if appErr != nil {
			response.Fail(
				ctx,
				appErr.Code,
				appErr.Msg,
			)
			return
		}

		logger :=
			logging.FromContext(
				ctx.Request.Context(),
			)

		logger.ErrorContext(
			ctx.Request.Context(),
			"http panic",

			slog.String(
				"method",
				ctx.Request.Method,
			),

			slog.String(
				"path",
				ctx.Request.URL.Path,
			),

			slog.Any(
				"error",
				err,
			),
		)

		response.Fail(
			ctx,
			response.CodeServerError,
			"internal server error",
		)

	}
}
