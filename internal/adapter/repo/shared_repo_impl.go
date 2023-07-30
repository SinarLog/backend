package repo

import (
	"context"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
)

type sharedRepo struct {
	db *gorm.DB
}

func NewSharedRepo(db *gorm.DB) *sharedRepo {
	return &sharedRepo{db}
}

func (repo *sharedRepo) GetRoleById(ctx context.Context, id string) (entity.Role, error) {
	var role entity.Role

	if err := repo.db.WithContext(ctx).Model(&role).First(&role, "id = ?", id).Error; err != nil {
		return role, err
	}

	return role, nil
}

func (repo *sharedRepo) GetJobById(ctx context.Context, id string) (entity.Job, error) {
	var job entity.Job

	if err := repo.db.WithContext(ctx).Model(&job).First(&job, "id = ?", id).Error; err != nil {
		return job, err
	}

	return job, nil
}
