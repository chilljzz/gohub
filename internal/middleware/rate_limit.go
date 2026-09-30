package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chilljzz/gohub/internal/logging"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var fixedWindowScript = redis.NewScript(
	`
		local current = redis.call("INCR",KEYS[1])

		if current == 1 
			then 
				redis.call("PEXPIRE",KEY[1],ARGV[1])
			end

		return current

	`,
)

type RateLimiter struct {
	client *redis.Client

	prefix string

	limit int64

	window time.Duration
}

func NewRateLimiter(
	client *redis.Client,
	prefix string,
	limit int64,
	window time.Duration,
) *RateLimiter {
	return &RateLimiter{
		client: client,
		prefix: prefix,
		limit:  limit,
		window: window,
	}
}

func (l *RateLimiter) allow(ctx context.Context, key string) (bool, error) {
	count, err := fixedWindowScript.Run(
		ctx,
		l.client,
		[]string{key},
		l.window.Milliseconds(),
	).Int64()

	if err != nil {
		return false, err
	}

	return count <= l.limit, nil

}

func (l *RateLimiter) ByIP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := fmt.Sprintf("%s:ip:%s", l.prefix, ctx.ClientIP())
		allowed, err := l.allow(ctx.Request.Context(), key)
		if err != nil {
			logger := logging.FromContext(
				ctx.Request.Context(),
			)

			logger.LogAttrs(
				ctx.Request.Context(),
				slog.LevelError,
				"rate limiter unavailable",

				slog.String("scope", "ip"),

				slog.Any("error", err),
			)
			ctx.Next()
			return
		}
		if !allowed {
			response.Fail(
				ctx,
				response.CodeTooManyRequests,
				"too many requests",
			)
			ctx.Abort()
			return
		}
	}
}

func (l *RateLimiter) ByUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		value, exists :=
			ctx.Get(
				"userID",
			)

		if !exists {

			response.Fail(
				ctx,
				response.CodeUnauthorized,
				"user identity not found",
			)

			ctx.Abort()
			return
		}

		userID, ok :=
			value.(uint)

		if !ok {
			return
		}

		key :=
			fmt.Sprintf(
				"%s:user:%d",
				l.prefix,
				userID,
			)

		allowed, err :=
			l.allow(
				ctx.Request.Context(),
				key,
			)

		if err != nil {

			slog.Error(
				"rate limiter unavailable",

				slog.String(
					"scope",
					"user",
				),

				slog.Any(
					"error",
					err,
				),
			)

			ctx.Next()
			return
		}

		if !allowed {

			response.Fail(
				ctx,
				response.CodeTooManyRequests,
				"too many requests",
			)

			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
