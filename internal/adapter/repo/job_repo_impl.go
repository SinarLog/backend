package repo

import (
	"context"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
)

type jobRepo struct {
	db *gorm.DB
}

func NewJobRepo(db *gorm.DB) *jobRepo {
	return &jobRepo{db}
}

func (repo *jobRepo) GetAllJobs(ctx context.Context) ([]entity.Job, error) {
	var jobs []entity.Job

	if err := repo.db.WithContext(ctx).Model(&entity.Job{}).Find(&jobs).Error; err != nil {
		return nil, err
	}

	return jobs, nil
}
