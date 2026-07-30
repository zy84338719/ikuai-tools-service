package dao

import (
	"context"
	"errors"

	"github.com/zy84338719/ikuai-tools-service/internal/repo/db"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
	"gorm.io/gorm"
)

// RouterRepository provides CRUD over the routers table.
type RouterRepository struct{}

func NewRouterRepository() *RouterRepository { return &RouterRepository{} }

func (r *RouterRepository) db() *gorm.DB { return db.GetDB() }

func (r *RouterRepository) List(ctx context.Context) ([]model.Router, error) {
	var rs []model.Router
	err := r.db().WithContext(ctx).Order("id ASC").Find(&rs).Error
	return rs, err
}

func (r *RouterRepository) ListEnabled(ctx context.Context) ([]model.Router, error) {
	var rs []model.Router
	err := r.db().WithContext(ctx).Where("status = ?", 1).Order("id ASC").Find(&rs).Error
	return rs, err
}

func (r *RouterRepository) GetByName(ctx context.Context, name string) (*model.Router, error) {
	var m model.Router
	if err := r.db().WithContext(ctx).Where("name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *RouterRepository) GetByID(ctx context.Context, id uint) (*model.Router, error) {
	var m model.Router
	if err := r.db().WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *RouterRepository) Create(ctx context.Context, m *model.Router) error {
	return r.db().WithContext(ctx).Create(m).Error
}

func (r *RouterRepository) Update(ctx context.Context, m *model.Router) error {
	return r.db().WithContext(ctx).Save(m).Error
}

func (r *RouterRepository) Delete(ctx context.Context, id uint) error {
	res := r.db().WithContext(ctx).Delete(&model.Router{}, id)
	if res.RowsAffected == 0 {
		return errors.New("router not found")
	}
	return res.Error
}
