package router

import (
	"github.com/chilljzz/gohub/internal/controller"
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
	api := r.Group("/api")
	userController := controller.NewUserController()
	user := api.Group("/users")
	{
		user.POST("/register", userController.Register)
	}
	return r
}
