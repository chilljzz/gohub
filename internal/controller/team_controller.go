package controller

import (
	"strconv"

	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/gin-gonic/gin"
)

type TeamController struct {
	teamService *service.TeamService
}

func NewTeamController(
	teamService *service.TeamService,
) *TeamController {
	return &TeamController{
		teamService: teamService,
	}
}

func (c *TeamController) Create(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}
	var req request.CreateTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}
	result, err := c.teamService.CreateTeam(userID, req)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	response.Success(ctx, result)
}

func (c *TeamController) ListMine(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}
	result, err := c.teamService.ListMyTeams(userID)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	response.Success(ctx, result)
}

func (c *TeamController) ListMembers(ctx *gin.Context) {
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
	members, err := c.teamService.ListMembers(userID, uint(teamID))
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return

	}

	response.Success(ctx, members)

}

func (c *TeamController) AddMember(ctx *gin.Context) {
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
	var req request.AddTeamMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {

		response.Fail(
			ctx,
			response.CodeInvalidParam,
			"invalid request parameters",
		)
		return
	}
	member, err := c.teamService.AddMember(userID, uint(teamID), req)
	if err != nil {

		ctx.Error(err)
		ctx.Abort()
		return

	}
	response.Success(
		ctx,
		member,
	)

}
