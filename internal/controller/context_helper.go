package controller

import "github.com/gin-gonic/gin"

func getCurrentUserID(ctx *gin.Context) (uint, bool) {
	value, exists := ctx.Get("userID")
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	if !ok {
		return 0, false
	}
	return userID, true
}
