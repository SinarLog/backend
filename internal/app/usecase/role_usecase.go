package usecase

import (
	"context"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/entity"
)

type roleUseCase struct {
	repo repo.IRoleRepo
}

func NewRoleUseCase(repo repo.IRoleRepo) *roleUseCase {
	return &roleUseCase{repo}
}

func (uc *roleUseCase) RetrieveRoles(ctx context.Context) ([]entity.Role, error) {
	roles, err := uc.repo.GetAllRoles(ctx)
	if err != nil {
		return nil, NewRepositoryError("Role", err)
	}

	return roles, nil
}
