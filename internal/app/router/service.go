package router

import (
	"context"
	"errors"
	"strings"

	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/dao"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
	"gorm.io/gorm"
)

type CreateRouterReq struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"`
	Insecure *bool  `json:"insecure"`
	Timeout  *int   `json:"timeout"`
	Comment  string `json:"comment"`
}

type UpdateRouterReq struct {
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"` // empty = keep existing
	Insecure *bool  `json:"insecure"`
	Timeout  *int   `json:"timeout"`
	Status   *int8  `json:"status"`
	Comment  string `json:"comment"`
}

type Service struct {
	repo *dao.RouterRepository
}

func NewService() *Service { return &Service{repo: dao.NewRouterRepository()} }

func (s *Service) List(ctx context.Context) ([]*model.RouterResp, error) {
	rs, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*model.RouterResp, len(rs))
	for i := range rs {
		out[i] = rs[i].ToResp()
	}
	return out, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*model.RouterResp, error) {
	r, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("router not found")
		}
		return nil, err
	}
	return r.ToResp(), nil
}

func (s *Service) Create(ctx context.Context, req *CreateRouterReq) (*model.RouterResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || req.BaseURL == "" || req.Token == "" {
		return nil, errors.New("name, base_url and token are required")
	}
	m := &model.Router{
		Name:    name,
		BaseURL: req.BaseURL,
		Token:   req.Token,
		Comment: req.Comment,
		Status:  1,
	}
	if req.Insecure != nil {
		m.Insecure = *req.Insecure
	} else {
		m.Insecure = true
	}
	if req.Timeout != nil && *req.Timeout > 0 {
		m.Timeout = *req.Timeout
	} else {
		m.Timeout = 30
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m.ToResp(), nil
}

func (s *Service) Update(ctx context.Context, name string, req *UpdateRouterReq) (*model.RouterResp, error) {
	r, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("router not found")
		}
		return nil, err
	}
	if req.BaseURL != "" {
		r.BaseURL = req.BaseURL
	}
	if req.Token != "" { // empty keeps the existing token
		r.Token = req.Token
	}
	if req.Insecure != nil {
		r.Insecure = *req.Insecure
	}
	if req.Timeout != nil && *req.Timeout > 0 {
		r.Timeout = *req.Timeout
	}
	if req.Status != nil {
		r.Status = *req.Status
	}
	if req.Comment != "" {
		r.Comment = req.Comment
	}
	if err := s.repo.Update(ctx, r); err != nil {
		return nil, err
	}
	return r.ToResp(), nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	r, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("router not found")
		}
		return err
	}
	return s.repo.Delete(ctx, r.ID)
}

// AllForManager returns enabled routers (with tokens) so the Manager registry
// can construct clients. This is the bridge between the CRUD layer and the
// connection layer.
func (s *Service) AllForManager(ctx context.Context) ([]model.Router, error) {
	return s.repo.ListEnabled(ctx)
}
