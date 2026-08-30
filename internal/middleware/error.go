package middleware

import (
	"log"
	"runtime/debug"

	"github.com/chilljzz/gohub/internal/response"
	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf(
					"http panic:methon=%s path=%s err=%v\n%s",
					ctx.Request.Method,
					ctx.Request.URL.Path,
					err,
					debug.Stack(),
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
			log.Printf(
				"error after response written: method=%s path=%s err=%v",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err,
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

		log.Printf(
			"http request error: method=%s path=%s err=%v",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			err,
		)

		response.Fail(
			ctx,
			response.CodeServerError,
			"internal server error",
		)

	}
}
