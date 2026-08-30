package service

import (
	"errors"
	"strings"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/request"
	"github.com/chilljzz/gohub/pkg/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("username already exists")
	ErrUsernameOrPassword = errors.New("username or password error")
	ErrUserNotFound       = errors.New("user not found")
	ErrNoProfileFields    = errors.New("no profile fields provided")
	ErrInvalidNickname    = errors.New("nickname cannot be empty")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) Register(req request.RegisterRequest) (*dto.RegisterResult, error) {
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
	return &dto.RegisterResult{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	}, nil
}

func (s *UserService) Login(req request.LoginRequest) (*dto.LoginResult, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUsernameOrPassword
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, ErrUsernameOrPassword
	}
	token, err := jwtutil.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResult{
		Token: token,
	}, nil
}

func (s *UserService) GetProfile(
	userID uint,
) (*dto.UserProfileResult, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return &dto.UserProfileResult{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:00:00"),
	}, nil
}

func (s *UserService) UpdateProfile(
	userID uint,
	req request.UpdateProfileRequest,
) (*dto.UserProfileResult, error) {
	updates := make(map[string]any)
	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if nickname == "" {
			return nil, ErrInvalidNickname
		}
		updates["nickname"] = nickname
	}

	if req.Avatar != nil {
		updates["avatar"] = strings.TrimSpace(*req.Avatar)
	}
	if req.Bio != nil {
		updates["Bio"] = strings.TrimSpace(*req.Bio)
	}
	if len(updates) == 0 {
		return nil, ErrNoProfileFields
	}
	if err := s.userRepo.UpdateProfile(userID, updates); err != nil {
		return nil, err
	}
	return s.GetProfile(userID)
}
