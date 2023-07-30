package repo

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
)

type credentialRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCredentialRepo(db *gorm.DB, rdis *redis.Client) *credentialRepo {
	return &credentialRepo{db: db, redis: rdis}
}

func (repo *credentialRepo) GetEmployeeByEmail(ctx context.Context, email string) (entity.Employee, error) {
	var employee entity.Employee

	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Preload("Role").
		Preload("Job").
		First(&employee, "email = ?", email).
		Error; err != nil {
		return employee, err
	}

	return employee, nil
}

func (repo *credentialRepo) GetEmployeeByIdV2(ctx context.Context, id string) (entity.Employee, error) {
	var employee entity.Employee

	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Preload("Role").
		First(&employee, "id = ?", id).Error; err != nil {
		return employee, err
	}

	return employee, nil
}

func (repo *credentialRepo) UpdateEmployeePassword(ctx context.Context, employee entity.Employee) error {
	if err := repo.db.
		WithContext(ctx).
		Model(&employee).
		Where("id = ?", employee.Id).
		Update("password", employee.Password).Error; err != nil {
		return err
	}

	return nil
}
