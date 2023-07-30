package repo

import (
	"context"

	"sinarlog.com/internal/entity/vo"
)

type IAnalyticsRepo interface {
	GetAttendanceAndLeaveQuotaAnalyticsByEmployeeId(ctx context.Context, employeeId string) (vo.BriefLeaveAndAttendanceAnalytics, error)
	GetDashboardAnalyticsHr(ctx context.Context) (vo.HrDashboardAnalytics, error)
}
