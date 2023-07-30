package repo

import (
	"context"

	"sinarlog.com/internal/entity"
)

type ISharedRepo interface {
	// Roles
	GetRoleById(ctx context.Context, id string) (entity.Role, error)

	// Jobs
	GetJobById(ctx context.Context, id string) (entity.Job, error)
}
