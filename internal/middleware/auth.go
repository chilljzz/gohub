package middleware

import (
	"strings"

	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(
				ctx,
				response.CodeUnauthorized,
				"authorization header is required",
			)
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			response.Fail(
				ctx,
				response.CodeUnauthorized,
				"invalid authorization header",
			)
			ctx.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if !strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") || tokenString == "" {

			response.Fail(
				ctx,
				response.CodeUnauthorized,
				"invalid authorization header",
			)
			ctx.Abort()
			return
		}

		claims, err := jwtutil.ParseToken(tokenString)
		if err != nil {
			response.Fail(
				ctx,
				response.CodeUnauthorized,
				"invalid or expired token",
			)
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)

		ctx.Next()

	}

}
