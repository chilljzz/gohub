package controller

import "github.com/gin-gonic/gin"

func abortWithError(
	ctx *gin.Context,
	err error,
) {
	_ = ctx.Error(err)
	ctx.Abort()
}
