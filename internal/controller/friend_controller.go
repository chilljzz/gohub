package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type FriendController struct {
	friendService *service.FriendService
}

func NewFriendController(
	friendService *service.FriendService,
) *FriendController {

	return &FriendController{
		friendService: friendService,
	}
}

func (c *FriendController) SendRequest(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}
	var req request.SendFriendRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}
	result, err := c.friendService.SendRequest(userID, req)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}
	response.Success(ctx, result)
}

func (c *FriendController) ListPendingRequests(
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

	result, err := c.friendService.ListPendingRequests(userID)
	if err != nil {
		response.Fail(
			ctx,
			response.CodeServerError,
			"server error",
		)
		return
	}
	response.Success(ctx, result)
}

func (c *FriendController) AcceptRequest(
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
	requestID64, err := strconv.ParseUint(
		ctx.Param("id"),
		10,
		64,
	)
	if err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid friend request id",
		)
		return
	}
	err = c.friendService.AcceptRequest(userID, uint(requestID64))
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.SuccessWithMsg(
		ctx,
		"friend request accepted",
		nil,
	)
}

func (c *FriendController) ListFriends(
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
	result, err := c.friendService.ListFriends(userID)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, result)

}
