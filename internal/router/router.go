package router

import (
	"github.com/chilljzz/gohub/internal/response"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{
			"msg": "pong",
		})
	})
	return r
}
