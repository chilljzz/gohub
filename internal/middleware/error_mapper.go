package middleware

import (
	"errors"

	"github.com/chilljzz/gohub/internal/apperror"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
)

func mapError(err error) *apperror.AppError {
	switch {

	case errors.Is(
		err,
		service.ErrNoProfileFields,
	):
		return &apperror.AppError{
			Code: response.CodeInvalidParam,
			Msg:  "no profile fieds provided",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrInvalidNickname,
	):
		return &apperror.AppError{
			Code: response.CodeInvalidParam,
			Msg:  "nickname cannot be empty",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrUserExists,
	):
		return &apperror.AppError{
			Code: response.CodeUserExists,
			Msg:  "username already exists",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrUsernameOrPassword,
	):
		return &apperror.AppError{
			Code: response.CodeUsernameOrPassword,
			Msg:  "username or password error",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrUserNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeUserNotFound,
			Msg:  "user not found",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrCannotAddSelf,
	):
		return &apperror.AppError{
			Code: response.CodeCannotAddSelf,
			Msg:  "cannot add youself",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrAlreadyFriends,
	):
		return &apperror.AppError{
			Code: response.CodeAlreadyFriends,
			Msg:  "already friend",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrFriendRequestNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeFriendRequestNotFound,
			Msg:  "friend request not found",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrNotRequestReceiver,
	):
		return &apperror.AppError{
			Code: response.CodeUnauthorized,
			Msg:  "not friend request receiver",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrFriendRequestProcessed,
	):
		return &apperror.AppError{
			Code: response.CodeFriendRequestProcessed,
			Msg:  "friend request already processed",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrFriendRequestExists,
	):
		return &apperror.AppError{
			Code: response.CodeFriendRequestExists,
			Msg:  "friend request already exists",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrAlreadyTeamMember,
	):
		return &apperror.AppError{
			Code: response.CodeAlreadyTeamMember,
			Msg:  "already team member",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrTeamNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeTeamNotFound,
			Msg:  "team not found",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrNotTeamOwner,
	):
		return &apperror.AppError{
			Code: response.CodeNotTeamOwner,
			Msg:  "not team owner",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrChannelExists,
	):
		return &apperror.AppError{
			Code: response.CodeChannelExists,
			Msg:  "channelExists",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrInvalidParam,
	):
		return &apperror.AppError{
			Code: response.CodeInvalidParam,
			Msg:  "invalid request parameters",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrChannelNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeChannelNotFound,
			Msg:  "channel not fount",
			Err:  err,
		}
	case errors.Is(
		err,
		service.ErrNotTeamMember,
	):
		return &apperror.AppError{
			Code: response.CodeNotTeamMember,
			Msg:  "not team member",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrMessageNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeMessageNotFound,
			Msg:  "message not found",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrMessageNotInChannel,
	):
		return &apperror.AppError{
			Code: response.CodeMessageNotInChannel,
			Msg:  "message does not belong to channel",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrConversationNotFound,
	):
		return &apperror.AppError{
			Code: response.CodeConversationNotFound,
			Msg:  "conversation not found",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrConversationCreateFailed,
	):
		return &apperror.AppError{
			Code: response.CodeFail,
			Msg:  "conversation create fail",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrCannotChatWithSelf,
	):
		return &apperror.AppError{
			Code: response.CodeCannotChatWithSelf,
			Msg:  "can not chat with self",
			Err:  err,
		}

	case errors.Is(
		err,
		service.ErrDirectChatRequiresFriend,
	):
		return &apperror.AppError{
			Code: response.CodeDirectChatRequiresFriend,
			Msg:  " not friends",
			Err:  err,
		}

	}

	return nil
}
