package usecase

import (
	"context"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type analyticsUseCase struct {
	analyicsRepo repo.IAnalyticsRepo
}

func NewAnalyticsUseCase(analyticsRepo repo.IAnalyticsRepo) *analyticsUseCase {
	return &analyticsUseCase{analyicsRepo: analyticsRepo}
}

func (uc *analyticsUseCase) RetrieveDashboardAnalyticsForEmployeeById(ctx context.Context, employeeId string) (vo.BriefLeaveAndAttendanceAnalytics, error) {
	anal, err := uc.analyicsRepo.GetAttendanceAndLeaveQuotaAnalyticsByEmployeeId(ctx, employeeId)
	if err != nil {
		return anal, NewRepositoryError("Analytics", err)
	}

	return anal, nil
}

func (uc *analyticsUseCase) RetrieveDashboardAnalyticsHr(ctx context.Context) (vo.HrDashboardAnalytics, error) {
	anal, err := uc.analyicsRepo.GetDashboardAnalyticsHr(ctx)
	if err != nil {
		return anal, NewRepositoryError("Analytics", err)
	}

	anal.Month = utils.GetStartOfTheMonth().Month().String()
	return anal, nil
}
