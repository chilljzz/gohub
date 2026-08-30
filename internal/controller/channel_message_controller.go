package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type ChannelMessageController struct {
	messageService *service.ChannelMessageService
}

func NewChannelMessageController(
	messageService *service.ChannelMessageService,
) *ChannelMessageController {
	return &ChannelMessageController{
		messageService: messageService,
	}
}

func (c *ChannelMessageController) ListRecent(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	channelID64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || channelID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid channel id",
		)
		return
	}
	limitText := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitText)
	if err != nil || limit <= 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid limit",
		)
		return
	}

	beforeID64, err := strconv.ParseUint(
		ctx.DefaultQuery("before_id", "0"),
		10,
		64,
	)
	if err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid before_id",
		)
		return
	}

	result, err := c.messageService.ListMessageBefore(
		userID, uint(channelID64), uint(beforeID64), limit)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.Success(ctx, result)
}

func (c *ChannelMessageController) Sync(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}
	channelID64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || channelID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid channel id",
		)
		return
	}
	channelID := uint(channelID64)

	var query request.SyncChannelMessageQuery

	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid query parameters",
		)
		return
	}
	result, err := c.messageService.SyncMessagesAfter(
		userID,
		channelID,
		query.AfterID,
		query.Limit,
	)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, result)
}
