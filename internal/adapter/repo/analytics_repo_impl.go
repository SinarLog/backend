package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type analyticsRepo struct {
	db *gorm.DB
}

func NewAnalyticsRepo(db *gorm.DB) *analyticsRepo {
	return &analyticsRepo{db}
}

func (repo *analyticsRepo) GetAttendanceAndLeaveQuotaAnalyticsByEmployeeId(ctx context.Context, employeeId string) (vo.BriefLeaveAndAttendanceAnalytics, error) {
	type aggregate1 struct {
		EmployeeId string `gorm:"type:uuid"`
		Lates      int
		Earlies    int
	}

	var agg1 aggregate1
	var res vo.BriefLeaveAndAttendanceAnalytics

	if err := repo.db.WithContext(ctx).Raw(`
	SELECT
		employee_id,
		(SELECT COUNT(*) AS lates FROM "attendances" WHERE late_clock_in = ? AND employee_id = ? AND clock_in_at BETWEEN ? AND ?),
		(SELECT COUNT(*) AS earlies FROM "attendances" WHERE early_clock_out = ? AND employee_id = ? AND clock_in_at BETWEEN ? AND ?)
	FROM "attendances"
	GROUP BY employee_id
	HAVING employee_id = ?
	`,
		true, employeeId, utils.GetStartOfTheMonth(), time.Now().In(utils.CURRENT_LOC),
		true, employeeId, utils.GetStartOfTheMonth(), time.Now().In(utils.CURRENT_LOC),
		employeeId,
	).Scan(&agg1).Error; err != nil {
		return res, err
	}

	res.LateClockIns = agg1.Lates
	res.EarlyClockOuts = agg1.Earlies

	row := repo.db.WithContext(ctx).Table("employee_leaves_quota").Where("employee_id = ?", employeeId).Select("yearly_count", "unpaid_count").Row()
	if err := row.Scan(&res.YearlyCount, &res.UnpaidCount); err != nil {
		return res, err
	}

	return res, nil
}

func (repo *analyticsRepo) GetDashboardAnalyticsHr(ctx context.Context) (vo.HrDashboardAnalytics, error) {
	var res vo.HrDashboardAnalytics

	stm := repo.db.WithContext(ctx).Session(&gorm.Session{PrepareStmt: true})

	// Total Employees
	if err := stm.Model(&entity.Employee{}).Count(&res.TotalEmployees).Error; err != nil {
		return res, err
	}

	// Lates and Earlies
	if err := stm.Raw(`
	SELECT
		(SELECT COUNT(*) AS late_clock_ins FROM "attendances" WHERE late_clock_in = ? AND clock_in_at BETWEEN ? AND ?),
		(SELECT COUNT(*) AS early_clock_outs   FROM "attendances" WHERE early_clock_out = ? AND clock_in_at BETWEEN ? AND ?)
	FROM "attendances"
	`,
		true, utils.GetStartOfTheMonth(), time.Now().In(utils.CURRENT_LOC),
		true, utils.GetStartOfTheMonth(), time.Now().In(utils.CURRENT_LOC),
	).Scan(&res).Error; err != nil {
		return res, err
	}

	// Unpaids
	if err := stm.Model(&entity.Leave{}).
		Where("approved_by_manager IS TRUE AND approved_by_hr IS TRUE").
		Where("created_at BETWEEN ? AND ?", utils.GetStartOfTheMonth(), utils.GetEndOfTheMonth()).
		Where("type = ?", entity.UNPAID.String()).
		Count(&res.ApprovedUnpaidLeaves).Error; err != nil {
		return res, err
	}

	// Annual and Marriages
	if err := stm.Model(&entity.Leave{}).
		Where("approved_by_manager IS TRUE AND approved_by_hr IS TRUE").
		Where("created_at BETWEEN ? AND ?", utils.GetStartOfTheMonth(), utils.GetEndOfTheMonth()).
		Where("(type = ? OR type = ?)", entity.MARRIAGE.String(), entity.ANNUAL.String()).
		Count(&res.ApprovedAnnualMarriageLeaves).Error; err != nil {
		return res, err
	}

	return res, nil
}
