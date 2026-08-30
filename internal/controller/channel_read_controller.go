package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type ChannelReadController struct {
	readService *service.ChannelReadService
}

func NewChannelReadController(
	readService *service.ChannelReadService,
) *ChannelReadController {
	return &ChannelReadController{
		readService: readService,
	}
}

func (c *ChannelReadController) MarkRead(
	ctx *gin.Context,
) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	channelID64, err := strconv.ParseUint(
		ctx.Param("id"),
		10,
		64,
	)
	if err != nil || channelID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid channel id",
		)
		return
	}
	channelID := uint(channelID64)

	var req request.MarkChannelReadRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid parameters",
		)
		return
	}

	err = c.readService.MarkRead(
		userID,
		channelID,
		req.MessageID,
	)

	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.SuccessWithMsg(ctx, "mark read success", req.MessageID)
}

func (c *ChannelReadController) UnreadCount(
	ctx *gin.Context,
) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	channelID64, err := strconv.ParseUint(
		ctx.Param("id"),
		10,
		64,
	)
	if err != nil || channelID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid channel id",
		)
		return
	}

	chanelID := uint(channelID64)
	result, err := c.readService.GetUnreadCount(
		userID,
		chanelID,
	)

	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, result)

}
