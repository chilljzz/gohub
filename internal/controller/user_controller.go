package controller

import (
	"errors"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req request.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, response.CodeInvalidParam, err.Error())
		return
	}
	result, err := c.userService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.Fail(ctx, response.CodeUserExists, "username already exists")
			return
		}
		response.Fail(ctx, response.CodeServerError, "server error")
		return
	}
	response.Success(ctx, result)

}
