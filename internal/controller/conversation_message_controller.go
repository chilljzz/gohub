package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type ConversationMessageController struct {
	service *service.ConversationMessageService
}

func NewConversationMessageController(
	service *service.ConversationMessageService,
) *ConversationMessageController {

	return &ConversationMessageController{
		service: service,
	}
}

func (c *ConversationMessageController) ListRecent(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	conversationID64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || conversationID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid conversation id",
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

	result, err := c.service.ListBefore(
		userID, uint(conversationID64), uint(beforeID64), limit)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.Success(ctx, result)
}

func (c *ConversationMessageController) Sync(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}
	conversationID64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || conversationID64 == 0 {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid conversation id",
		)
		return
	}
	conversationID := uint(conversationID64)

	var query request.SyncChannelMessageQuery

	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid query parameters",
		)
		return
	}
	result, err := c.service.SyncAfter(
		userID,
		conversationID,
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
