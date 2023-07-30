package usecase

import (
	"context"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/entity"
)

type jobUseCase struct {
	repo repo.IJobRepo
}

func NewJobUseCase(repo repo.IJobRepo) *jobUseCase {
	return &jobUseCase{repo}
}

func (uc *jobUseCase) RetrieveJobs(ctx context.Context) ([]entity.Job, error) {
	jobs, err := uc.repo.GetAllJobs(ctx)
	if err != nil {
		return nil, NewRepositoryError("Job", err)
	}

	return jobs, nil
}