package user

import (
	"context"
	"errors"

	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/dao"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUserReq struct {
	Username string
	Email    string
	Password string
	Nickname string
}

type UpdateUserReq struct {
	Nickname string
	Avatar   string
	Status   *int8
}

type Service struct {
	repo *dao.UserRepository
}

func NewService() *Service {
	return &Service{
		repo: dao.NewUserRepository(),
	}
}

func (s *Service) Create(ctx context.Context, req *CreateUserReq) (*model.UserResp, error) {
	existing, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	existing, err = s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResp(), nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*model.UserResp, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user.ToResp(), nil
}

func (s *Service) Update(ctx context.Context, id uint, req *UpdateUserReq) (*model.UserResp, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResp(), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.UserResp, int64, error) {
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]*model.UserResp, len(users))
	for i, user := range users {
		resps[i] = user.ToResp()
	}

	return resps, total, nil
}
