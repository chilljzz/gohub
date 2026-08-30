package controller

import (
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

func (c *UserController) Me(ctx *gin.Context) {
	userID, exists := getCurrentUserID(ctx)
	if !exists {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	result, err := c.userService.GetProfile(userID)
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.Success(ctx, result)
}

func (c *UserController) UpdateMe(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
	}
	var req request.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invaild request parameters",
		)
		return
	}
	result, err := c.userService.UpdateProfile(userID, req)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}
	response.Success(ctx, result)

}

func (c *UserController) Register(ctx *gin.Context) {
	var req request.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, response.CodeInvalidParam, err.Error())
		return
	}
	result, err := c.userService.Register(req)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.Success(ctx, result)

}

func (c *UserController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}
	result, err := c.userService.Login(req)
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.Success(ctx, result)
}
