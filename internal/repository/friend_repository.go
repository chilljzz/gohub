package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

var ErrFriendRequestStateChanged = errors.New(
	"friend request state changed",
)

type FriendRepository struct{}

func NewFriendREpository() *FriendRepository {
	return &FriendRepository{}
}

func (r *FriendRepository) IsFriend(
	userID uint,
	friendID uint,
) (bool, error) {
	var count int64
	err := database.DB.
		Model(&model.Friendship{}).
		Where(
			"user_id = ? AND friend_id = ?",
			userID,
			friendID,
		).Count(&count).Error

	return count > 0, err
}

func (r *FriendRepository) FindPendingBetween(
	userID uint,
	targetID uint,
) (*model.FriendRequest, error) {
	var friendRequest model.FriendRequest
	err := database.DB.
		Where(
			`status = ? AND (
			(sender_id = ? AND receiver_id = ?)
			OR
			(sender_id = ? AND receiver_id = ?)
			)`,
			model.FriendRequestPending,
			userID,
			targetID,
			targetID,
			userID,
		).
		First(&friendRequest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &friendRequest, nil
}

func (r *FriendRepository) CreateRequest(
	friendRequest *model.FriendRequest,
) error {
	return database.DB.Create(friendRequest).Error
}

func (r *FriendRepository) ListPendingRequests(
	receiverID uint,
) ([]model.FriendRequest, error) {
	var friednRequests []model.FriendRequest

	err := database.DB.
		Preload("Sender").
		Where(
			"receiver_id = ? AND status = ?",
			receiverID,
			model.FriendRequestPending,
		).
		Order("id DESC").
		Find(&friednRequests).Error

	return friednRequests, err
}

func (r *FriendRepository) FindRequestByID(
	requestID uint,
) (*model.FriendRequest, error) {
	var friendRequest model.FriendRequest

	err := database.DB.
		First(&friendRequest, requestID).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &friendRequest, nil
}

func (r *FriendRepository) AcceptRequest(
	friendRequest *model.FriendRequest,
) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.FriendRequest{}).
			Where(
				"id = ? AND receiver_id = ? AND status = ?",
				friendRequest.ID,
				friendRequest.ReceiverID,
				model.FriendRequestPending,
			).
			Update("status", model.FriendRequestAccepted)

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrFriendRequestStateChanged
		}

		firstRelation := model.Friendship{
			UserID:   friendRequest.SenderID,
			FriendID: friendRequest.ReceiverID,
		}

		if err := tx.
			Where(
				"user_id = ? AND friend_id = ?",
				firstRelation.UserID,
				firstRelation.FriendID,
			).FirstOrCreate(&firstRelation).Error; err != nil {
			return err
		}

		sencondRelation := model.Friendship{
			UserID:   friendRequest.ReceiverID,
			FriendID: friendRequest.SenderID,
		}
		if err := tx.Where(
			"user_id = ? AND friend_id = ?",
			sencondRelation.UserID,
			sencondRelation.FriendID,
		).FirstOrCreate(&sencondRelation).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *FriendRepository) ListFriends(
	userID uint,
) ([]model.User, error) {
	var users []model.User
	err := database.DB.
		Table("users").
		Select("users.*").
		Joins(
			"JOIN friendships ON friendships.friend_id = users.id",
		).
		Where("friendships.user_id = ?", userID).
		Order("friendships.id DESC").
		Find(&users).
		Error

	return users, err
}
