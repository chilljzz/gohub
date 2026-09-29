package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/chilljzz/gohub/internal/logging"
	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := newRequestID()

		ctx.Header("X-REquest-ID", requestID)

		requestCtx := logging.WithRequestID(
			ctx.Request.Context(),
			requestID,
		)

		ctx.Request = ctx.Request.WithContext(requestCtx)

		ctx.Next()
	}
}

func newRequestID() string {

	var date [16]byte

	if _, err := rand.Read(date[:]); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(date[:])
}
