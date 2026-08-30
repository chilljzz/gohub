package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type ChannelController struct {
	channelService *service.ChannelService
}

func NewChannelController(
	channelService *service.ChannelService,
) *ChannelController {
	return &ChannelController{
		channelService: channelService,
	}
}

func (c *ChannelController) Create(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}

	var req request.CreateChannelRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}

	result, err := c.channelService.CreateChannel(userID, uint(teamID), req)

	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, result)

}

func (c *ChannelController) List(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}

	result, err := c.channelService.ListChannels(userID, uint(teamID))
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, result)
}
