package repo

import (
	"context"

	"sinarlog.com/internal/entity"
)

type IJobRepo interface {
	GetAllJobs(ctx context.Context) ([]entity.Job, error)
}