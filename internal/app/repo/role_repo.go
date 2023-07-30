package repo

import (
	"context"

	"sinarlog.com/internal/entity"
)

type IRoleRepo interface {
	GetAllRoles(ctx context.Context) ([]entity.Role, error)
}
