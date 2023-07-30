package repo

import (
	"context"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
)

type roleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *roleRepo {
	return &roleRepo{db}
}

func (repo *roleRepo) GetAllRoles(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role

	if err := repo.db.WithContext(ctx).Model(entity.Role{}).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}
