package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	CodeSuccess = 0
	CodeFail    = 1

	CodeInvalidParam = 4000
	CodeUnauthorized = 4010
	CodeForbidden    = 4030
	CodeServerError  = 5000

	CodeUserExists         = 4001
	CodeUsernameOrPassword = 4002
	CodeUserNotFound       = 4003

	CodeCannotAddSelf          = 4100
	CodeAlreadyFriends         = 4101
	CodeFriendRequestExists    = 4102
	CodeFriendRequestNotFound  = 4103
	CodeFriendRequestProcessed = 4104

	CodeTeamNotFound      = 6001
	CodeNotTeamMember     = 6002
	CodeNotTeamOwner      = 6003
	CodeAlreadyTeamMember = 6004

	CodeMessageNotFound     = 7001
	CodeMessageNotInChannel = 7002

	CodeChannelNotFound = 6101
	CodeChannelExists   = 6102
)

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "SUCCESS",
		Data: data,
	})
}

func SuccessWithMsg(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

func Fail(ctx *gin.Context, code int, msg string) {
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
