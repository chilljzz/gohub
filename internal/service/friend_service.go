package service

import (
	"errors"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/request"
)

var (
	ErrCannotAddSelf          = errors.New("cannot add yourself")
	ErrAlreadyFriends         = errors.New("already friends")
	ErrFriendRequestExists    = errors.New("friend request already exists")
	ErrFriendRequestNotFound  = errors.New("friend request not found")
	ErrNotRequestReceiver     = errors.New("not friend request receiver")
	ErrFriendRequestProcessed = errors.New("friend request already processed")
)

type FriendService struct {
	friendRepo *repository.FriendRepository
	userRepo   *repository.UserRepository
}

func NewFriendService(
	friendRepo *repository.FriendRepository,
	userRepo *repository.UserRepository,
) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

func (s *FriendService) SendRequest(
	senderID uint,
	req request.SendFriendRequest,
) (*dto.FriendRequestCreatedResult, error) {
	receiver, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if receiver == nil {
		return nil, ErrUserNotFound
	}
	if senderID == receiver.ID {
		return nil, ErrCannotAddSelf
	}

	isFriend, err := s.friendRepo.IsFriend(
		senderID,
		receiver.ID,
	)
	if err != nil {
		return nil, err
	}

	if isFriend {
		return nil, ErrAlreadyFriends
	}

	pendingRequest, err := s.friendRepo.FindPendingBetween(
		senderID,
		receiver.ID,
	)
	if err != nil {
		return nil, err
	}

	if pendingRequest != nil {
		return nil, ErrFriendRequestExists
	}

	friendRequest := &model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiver.ID,
		Status:     model.FriendRequestPending,
	}
	if err := s.friendRepo.CreateRequest(friendRequest); err != nil {
		return nil, err
	}
	return &dto.FriendRequestCreatedResult{
		ID: friendRequest.ID,
		Receiver: dto.UserBridfResult{
			ID:       receiver.ID,
			Username: receiver.Username,
			Nickname: receiver.Nickname,
			Avatar:   receiver.Avatar,
		},
		Status:    friendRequest.Status,
		CreatedAt: friendRequest.CreatedAt,
	}, nil
}

func (s *FriendService) ListPendingRequests(
	userID uint,
) ([]dto.FriendRequestResult, error) {
	friendRequest, err := s.friendRepo.ListPendingRequests(userID)
	if err != nil {
		return nil, err
	}
	results := make([]dto.FriendRequestResult, 0, len(friendRequest))
	for _, item := range friendRequest {
		results = append(results, dto.FriendRequestResult{
			ID: item.ID,
			Sender: dto.UserBridfResult{
				ID:       item.Sender.ID,
				Username: item.Sender.Username,
				Nickname: item.Sender.Nickname,
				Avatar:   item.Sender.Avatar,
			},
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		})
	}
	return results, nil
}

func (s *FriendService) AcceptRequest(
	userID uint,
	requestID uint,
) error {
	friendRequset, err := s.friendRepo.FindRequestByID(requestID)
	if err != nil {
		return nil
	}
	if friendRequset == nil {
		return ErrFriendRequestNotFound
	}
	if friendRequset.ReceiverID != userID {
		return ErrNotRequestReceiver
	}

	if friendRequset.Status != model.FriendRequestPending {
		return ErrFriendRequestProcessed
	}

	err = s.friendRepo.AcceptRequest(friendRequset)
	if errors.Is(err, repository.ErrFriendRequestStateChanged) {
		return ErrFriendRequestProcessed
	}
	return err
}

func (s *FriendService) ListFriends(
	userID uint,
) ([]dto.UserBridfResult, error) {
	users, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}

	results := make([]dto.UserBridfResult, 0, len(users))
	for _, user := range users {
		results = append(results, dto.UserBridfResult{
			ID:       userID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}
	return results, nil
}
