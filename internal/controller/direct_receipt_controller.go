package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type DirectReceiptController struct {
	directReceiptService *service.DirectReceiptService
}

func NewDirectReceiptController(
	directReceiptService *service.DirectReceiptService,
) *DirectReceiptController {
	return &DirectReceiptController{
		directReceiptService: directReceiptService,
	}
}

func (c *DirectReceiptController) ReceiptState(ctx *gin.Context) {
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

	result, err := c.directReceiptService.GetPeerState(userID, uint(conversationID64))

	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.Success(ctx, result)
}
