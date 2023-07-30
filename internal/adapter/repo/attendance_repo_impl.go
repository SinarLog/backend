package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type attendanceRepo struct {
	db   *gorm.DB
	rdis *redis.Client
}

func NewAttendanceRepo(db *gorm.DB, rdis *redis.Client) *attendanceRepo {
	return &attendanceRepo{db: db, rdis: rdis}
}

func (repo *attendanceRepo) EmployeeHasClockedInToday(ctx context.Context, employeeId string) (bool, error) {
	// If count is 0, this means that the employee has not clocked in today.
	var count int64

	if err := repo.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("employee_id = ?", employeeId).
		Where("clock_in_at BETWEEN ? AND ?",
			utils.GetStartOfDay(),
			utils.GetEndOfDay()).
		Count(&count).
		Error; err != nil {
		return false, err
	}

	return count != 0, nil
}

func (repo *attendanceRepo) CreateNewAttendance(ctx context.Context, attendance entity.Attendance) error {
	if err := repo.db.WithContext(ctx).
		Model(&attendance).
		Create(&attendance).
		Error; err != nil {
		return err
	}

	return nil
}

func (repo *attendanceRepo) SaveClockInOTPTimestamp(ctx context.Context, emplooyeeId string, timestamp int64, exp time.Duration) error {
	if err := repo.rdis.Set(
		ctx,
		fmt.Sprintf("%s:clockIn", emplooyeeId),
		timestamp,
		exp,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (repo *attendanceRepo) GetClockInOTPTimestamp(ctx context.Context, employeeId string) (int64, error) {
	timestamp, err := repo.rdis.Get(ctx, fmt.Sprintf("%s:clockIn", employeeId)).Int64()
	if err != nil {
		if err == redis.Nil {
			return timestamp, fmt.Errorf("OTP not found")
		}

		return timestamp, err
	}

	return timestamp, nil
}

// GetTodaysAttendanceByEmployeeId queries the employee's today's attendance
func (repo *attendanceRepo) GetTodaysAttendanceByEmployeeId(ctx context.Context, employeeId string) (entity.Attendance, error) {
	var attendance entity.Attendance

	if err := repo.db.WithContext(ctx).
		Model(&attendance).
		Where("employee_id = ?", employeeId).
		Where("clock_in_at BETWEEN ? AND ?", utils.GetStartOfDay(), utils.GetEndOfDay()).
		Order("created_at DESC").
		Limit(1).
		Find(&attendance).Error; err != nil {
		return entity.Attendance{}, err
	}

	return attendance, nil
}

// EmployeeHasActiveAttendance checks whether an employee has an active attendance for today.
// An active attendance is defined by clock_in_at value is not null or empty time and is between
// the start and end of day, and clock_out_at value is null or an empty time.
func (repo *attendanceRepo) EmployeeHasActiveAttendance(ctx context.Context, employeeId string) (bool, error) {
	// If count = 0, then there are no active attendance.
	var count int64
	var emptyTime time.Time

	if err := repo.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("employee_id = ?", employeeId).
		Where("clock_in_at BETWEEN ? AND ?", utils.GetStartOfDay(), utils.GetEndOfDay()).
		Where("done_for_the_day = ?", false).
		Where("closed_automatically IS NULL").
		Where("clock_out_at NOT BETWEEN ? AND ? OR clock_out_at = ?",
			utils.GetStartOfDay(),
			utils.GetEndOfDay(),
			emptyTime,
		).Count(&count).Error; err != nil {
		return false, err
	}

	return count != 0, nil
}

// GetActiveAttendanceByEmployeeId queries the employee's active attendance. Different from
// GetTodaysAttendanceByEmployeeId, where it queries the attendance today whether active
// or not active or even not existing. Here, it requires the attendance to be existing and
// active.
func (repo *attendanceRepo) GetActiveAttendanceByEmployeeId(ctx context.Context, employeeId string) (entity.Attendance, error) {
	var attendance entity.Attendance

	if err := repo.db.WithContext(ctx).
		Model(&attendance).
		Where("employee_id = ?", employeeId).
		Where("clock_in_at BETWEEN ? AND ?", utils.GetStartOfDay(), utils.GetEndOfDay()).
		Where("clock_out_at NOT BETWEEN ? AND ? OR clock_out_at = ?",
			utils.GetStartOfDay(),
			utils.GetEndOfDay(),
			time.Time{},
		).
		Order("created_at DESC").
		First(&attendance).Error; err != nil {
		return entity.Attendance{}, err
	}

	return attendance, nil
}

func (repo *attendanceRepo) SumWeeklyOvertimeDurationByEmployeeId(ctx context.Context, employeeId string) (int, error) {
	// Return of the DB could be null
	var sum *int
	var truee bool = true

	if err := repo.db.WithContext(ctx).
		Raw(`
		SELECT SUM(duration) 
		FROM "overtimes" 
		INNER JOIN "attendances" 
		ON 
			"attendances"."id" = "overtimes"."attendance_id" 
			AND 
			"attendances"."employee_id" = ? 
		WHERE 
			"overtimes"."created_at" BETWEEN ? AND ? 
			AND
			(
				(
				"overtimes"."approved_by_manager" IS NULL
				OR
				"overtimes"."approved_by_manager" = ?
				)
				AND 
				"overtimes"."closed_automatically" IS NULL
			)`,
			employeeId,
			utils.GetStartOfTheWeekFromToday(),
			utils.GetEndOfWeekdayFromToday(),
			&truee,
		).Scan(&sum).Error; err != nil {
		return 0, err
	}

	if sum == nil {
		return 0, nil
	}

	return *sum, nil
}

// CloseAttendance closes an active attendance an save all of its associations state.
func (repo *attendanceRepo) CloseAttendance(ctx context.Context, attendance entity.Attendance) error {
	if err := repo.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Updates(&attendance).
		Error; err != nil {
		return err
	}

	return nil
}

func (repo *attendanceRepo) GetOvertimeById(ctx context.Context, id string) (entity.Overtime, error) {
	var overtime entity.Overtime

	if err := repo.db.WithContext(ctx).
		Model(&overtime).
		Preload("Attendance.Employee").
		Preload("Manager").
		First(&overtime, "id = ?", id).
		Error; err != nil {
		return overtime, err
	}

	return overtime, nil
}

func (repo *attendanceRepo) GetIncomingOvertimeSubmissionsForManager(ctx context.Context, managerId string, q vo.IncomingOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()

	var overtimes []entity.Overtime
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Overtime{}).
		Where("approved_by_manager IS NULL").
		Where("action_by_manager_at IS NULL").
		Where("closed_automatically IS NULL").
		Where(`"overtimes"."manager_id" = ?`, managerId)

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "attendances" ON "attendances"."id" = "overtimes"."attendance_id" INNER JOIN "employees" ON "employees"."id" = "attendances"."employee_id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if err := t.
		Preload("Attendance.Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&overtimes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return overtimes, pquery.Compress(count), nil
}

func (repo *attendanceRepo) SaveProcessedOvertimeSubmissionByManager(ctx context.Context, overtime entity.Overtime) error {
	if err := repo.db.
		WithContext(ctx).
		Model(&overtime).
		Omit("attendance_id", "duration", "reason", "manager_id", "Attendance").
		Updates(entity.Overtime{
			ApprovedByManager: overtime.ApprovedByManager,
			RejectionReason:   overtime.RejectionReason,
			ActionByManagerAt: overtime.ActionByManagerAt,
		}).
		Error; err != nil {
		return err
	}

	return nil
}

func (repo *attendanceRepo) GetMyOvertimeSubmissions(ctx context.Context, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()

	var overtimes []entity.Overtime
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Overtime{}).Joins(`INNER JOIN "attendances" ON "attendances"."id" = "overtimes"."attendance_id" AND "attendances"."employee_id" = ?`, employeeId)

	switch strings.ToLower(q.Status) {
	case "pending":
		t = t.Where(`approved_by_manager IS NULL AND "overtimes"."closed_automatically" IS NULL`)
	case "approved":
		t = t.Where(`approved_by_manager IS NOT NULL AND approved_by_manager IS TRUE AND "overtimes"."closed_automatically" IS NULL`)
	case "rejected":
		t = t.Where(`approved_by_manager IS NOT NULL AND approved_by_manager IS FALSE AND "overtimes"."closed_automatically" IS NULL`)
	case "closed":
		t = t.Where(`"overtimes"."closed_automatically" IS NOT NULL AND "overtimes"."closed_automatically" IS TRUE`)
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"overtimes"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "overtimes"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "overtimes"."created_at") = ?`, tquery.Year)
	}

	if err := t.Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&overtimes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return overtimes, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetOvertimeSubmissionHistoryForManager(ctx context.Context, managerId string, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()

	var overtimes []entity.Overtime
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Overtime{}).Where(`"overtimes"."manager_id" = ?`, managerId)

	switch strings.ToLower(q.Status) {
	case "approved":
		t = t.Where("approved_by_manager IS TRUE AND closed_automatically IS NULL")
	case "rejected":
		t = t.Where("approved_by_manager IS FALSE AND closed_automatically IS NULL")
	case "closed":
		t = t.Where("closed_automatically IS NOT NULL AND closed_automatically IS TRUE")
	default:
		t = t.Where("approved_by_manager IS NOT NULL")
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"overtimes"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "overtimes"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "overtimes"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "attendances" ON "attendances"."id" = "overtimes"."attendance_id" INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if err := t.Preload("Attendance.Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&overtimes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return overtimes, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetOvertimeSubmissionHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()

	var overtimes []entity.Overtime
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Overtime{})

	switch strings.ToLower(q.Status) {
	case "pending":
		t = t.Where("approved_by_manager IS NULL AND closed_automatically IS NULL")
	case "approved":
		t = t.Where("approved_by_manager IS TRUE AND closed_automatically IS NULL")
	case "rejected":
		t = t.Where("approved_by_manager IS FALSE AND closed_automatically IS NULL")
	case "closed":
		t = t.Where("closed_automatically IS TRUE AND approved_by_manager IS NULL")
	default:
		t = t.Where("1=?", 1)
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"overtimes"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "overtimes"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "overtimes"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "attendances" ON "attendances"."id" = "overtimes"."attendance_id" INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if err := t.Preload("Attendance.Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&overtimes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return overtimes, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetMyAttendancesHistory(ctx context.Context, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	pquery := q.CommonQuery.Pagination.MustExtract()
	tquery, _ := q.CommonQuery.TimeQuery.Extract()

	var attendances []entity.Attendance
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Attendance{}).Where("employee_id = ?", employeeId).Where("done_for_the_day IS TRUE")

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"attendances"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "attendances"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "attendances"."created_at") = ?`, tquery.Year)
	}

	if q.Late {
		t = t.Where(`"attendances"."late_clock_in" IS TRUE`)
	}

	if q.Early {
		t = t.Where(`"attendances"."early_clock_out" IS TRUE`)
	}

	if q.Closed {
		t = t.Where(`"attendances"."closed_automatically" IS TRUE`)
	}

	if err := t.Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&attendances).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return attendances, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetStaffsAttendancesHistory(ctx context.Context, managerId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	pquery := q.CommonQuery.Pagination.MustExtract()
	tquery, _ := q.CommonQuery.TimeQuery.Extract()

	var attendances []entity.Attendance
	var count int64

	t := repo.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("done_for_the_day IS TRUE")

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"attendances"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "attendances"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "attendances"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."manager_id" = ? AND "employees"."full_name" ILIKE ?`, managerId, utils.ToPatternMatching(q.Name))
	} else {
		t = t.Joins(`INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."manager_id" = ?`, managerId)
	}

	if q.Late {
		t = t.Where(`"attendances"."late_clock_in" IS TRUE`)
	}

	if q.Early {
		t = t.Where(`"attendances"."early_clock_out" IS TRUE`)
	}

	if q.Closed {
		t = t.Where(`"attendances"."closed_automatically" IS TRUE`)
	}

	if err := t.Preload("Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&attendances).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return attendances, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetEmployeesAttendanceHistory(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	pquery := q.CommonQuery.Pagination.MustExtract()
	tquery, _ := q.CommonQuery.TimeQuery.Extract()

	var attendances []entity.Attendance
	var count int64

	t := repo.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("done_for_the_day IS TRUE")

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"attendances"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "attendances"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "attendances"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if q.Late {
		t = t.Where(`"attendances"."late_clock_in" IS TRUE`)
	}

	if q.Early {
		t = t.Where(`"attendances"."early_clock_out" IS TRUE`)
	}

	if q.Closed {
		t = t.Where(`"attendances"."closed_automatically" IS TRUE`)
	}

	if err := t.Preload("Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&attendances).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return attendances, pquery.Compress(count), nil
}

func (repo *attendanceRepo) GetEmployeesTodaysAttendances(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	pquery := q.CommonQuery.Pagination.MustExtract()

	var attendances []entity.Attendance
	var count int64

	t := repo.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where(gorm.Expr(`"attendances"."clock_in_at" BETWEEN ? AND ?`, utils.GetStartOfDay(), utils.GetEndOfDay()))

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "attendances"."employee_id" = "employees"."id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if q.Late {
		t = t.Where(`"attendances"."late_clock_in" IS TRUE`)
	}

	if q.Early {
		t = t.Where(`"attendances"."early_clock_out" IS TRUE`)
	}

	if q.Closed {
		t = t.Where(`"attendances"."closed_automatically" IS TRUE`)
	}

	if err := t.Preload("Employee.Job").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&attendances).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return attendances, pquery.Compress(count), nil
}
