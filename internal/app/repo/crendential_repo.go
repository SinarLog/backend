package repo

import (
	"context"

	"sinarlog.com/internal/entity"
)

type ICredentialRepo interface {
	GetEmployeeByEmail(ctx context.Context, email string) (entity.Employee, error)
	GetEmployeeByIdV2(ctx context.Context, id string) (entity.Employee, error)
	UpdateEmployeePassword(ctx context.Context, employee entity.Employee) error
}
