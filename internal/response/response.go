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

	CodeTooManyRequests = 4290

	CodeTeamNotFound      = 6001
	CodeNotTeamMember     = 6002
	CodeNotTeamOwner      = 6003
	CodeAlreadyTeamMember = 6004

	CodeMessageNotFound     = 7001
	CodeMessageNotInChannel = 7002

	CodeChannelNotFound = 6101
	CodeChannelExists   = 6102

	// Conversation
	CodeConversationNotFound     = 42001
	CodeCannotChatWithSelf       = 42002
	CodeDirectChatRequiresFriend = 42003
	CodeNotConversationMember    = 42004
	CodeConversationConflict     = 42005
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
	ctx.JSON(httpStatusForCode(code), Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func httpStatusForCode(
	code int,
) int {

	switch code {

	case CodeUnauthorized,
		CodeUsernameOrPassword:

		return http.StatusUnauthorized

	case CodeForbidden,
		CodeNotTeamMember,
		CodeNotTeamOwner,
		CodeNotConversationMember,
		CodeDirectChatRequiresFriend:

		return http.StatusForbidden

	case CodeUserNotFound,
		CodeFriendRequestNotFound,
		CodeTeamNotFound,
		CodeChannelNotFound,
		CodeMessageNotFound,
		CodeConversationNotFound:

		return http.StatusNotFound

	case CodeUserExists,
		CodeAlreadyFriends,
		CodeFriendRequestExists,
		CodeFriendRequestProcessed,
		CodeAlreadyTeamMember,
		CodeChannelExists,
		CodeConversationConflict:

		return http.StatusConflict

	case CodeServerError:

		return http.StatusInternalServerError

	case CodeTooManyRequests:

		return http.StatusTooManyRequests

	default:

		return http.StatusBadRequest
	}
}
