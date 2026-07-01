package service

import (
	"errors"

	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/request"
	"golang.org/x/crypto/bcrypt"
)

type RegisterResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

var ErrUserExists = errors.New("username already exists")

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

type RegisterRequest struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

func (s *UserService) Register(req request.RegisterRequest) (*RegisterResult, error) {
	existUser, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existUser != nil {
		return nil, ErrUserExists
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	user := &model.User{
		Username: req.Username,
		Password: string(hashPassword),
		Nickname: nickname,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return &RegisterResult{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	}, nil
}
