package controller

import (
	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type ConversationController struct {
	conversationService *service.ConversationService
}

func NewConversationController(
	conversationService *service.ConversationService,
) *ConversationController {
	return &ConversationController{
		conversationService: conversationService,
	}
}

func (c *ConversationController) CreateDirect(
	ctx *gin.Context,
) {
	userID, ok := getCurrentUserID(ctx)

	if !ok {
		return
	}

	var req request.CreateDirectConversationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			400,
			"invalid request",
		)
		return
	}

	conversation, err :=
		c.conversationService.
			GetOrCreateDirectConversation(
				userID,
				req.UserID,
			)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.Success(
		ctx,
		conversation,
	)
}

func (c *ConversationController) List(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)

	if !ok {

		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)

		return
	}

	var query request.ListConversationQuery

	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid query parameters",
		)
		return
	}

	result, err := c.conversationService.ListConversations(
		userID,
		query.Limit,
		query.Cursor,
	)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.Success(
		ctx,
		result,
	)
}
