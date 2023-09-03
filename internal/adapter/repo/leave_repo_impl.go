package repo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type leaveRepo struct {
	db *gorm.DB
}

func NewLeaveRepo(db *gorm.DB) *leaveRepo {
	return &leaveRepo{db}
}

/*
*********************************
ACTOR: STAFF and MANAGER
*********************************
*/
func (repo *leaveRepo) EmployeeIsOnLeaveToday(ctx context.Context, employeeId string) (bool, error) {
	var count int64

	truee := true
	now := time.Now().In(utils.CURRENT_LOC)
	if err := repo.db.WithContext(ctx).
		Model(&entity.Leave{}).
		Where("employee_id = ?", employeeId).
		Where(`? BETWEEN "leaves"."from" AND "leaves"."to"`, now).
		Where("approved_by_manager IS NOT NULL").
		Where("approved_by_manager = ?", &truee).
		Where("approved_by_hr IS NOT NULL").
		Where("approved_by_hr = ?", &truee).
		Count(&count).
		Error; err != nil {
		return false, err
	}

	return count != 0, nil
}

func (repo *leaveRepo) GetMyLeaveRequestsList(ctx context.Context, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	q.Pagination.Sort = "DESC"
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()
	status := q.Status

	var leaves []entity.Leave
	var count int64

	// Old query
	// Where("approved_by_hr IS NULL OR approved_by_manager IS NULL OR EXTRACT('month' FROM DATE_TRUNC('month', now() - created_at)) <= ?", 1)
	t := repo.db.WithContext(ctx).Model(&entity.Leave{}).Where("employee_id = ?", employeeId)

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."created_at") = ?`, tquery.Year)
	}

	switch status {
	case "pending":
		t = t.Where(`
		(
			approved_by_manager IS TRUE
			OR
			approved_by_manager IS NULL
		)
		AND
		approved_by_hr IS NULL
		AND
		closed_automatically IS NULL
		`)
	case "approved":
		t = t.Where("approved_by_manager IS TRUE AND approved_by_hr IS TRUE AND closed_automatically IS NULL")
	case "rejected":
		t = t.Where("approved_by_manager IS NOT NULL AND (approved_by_manager IS FALSE OR approved_by_hr IS FALSE) AND closed_automatically IS NULL")
	case "closed":
		t = t.Where("closed_automatically IS TRUE")
	}

	if err := t.Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

func (repo *leaveRepo) CheckDateAvailability(ctx context.Context, leave entity.Leave) (bool, error) {
	var count int64
	var truee bool = true

	if err := repo.db.WithContext(ctx).Model(&leave).
		Where(`
			"employee_id" = ?
			AND
			(
				("leaves"."from" BETWEEN ? AND ?)
				OR
				("leaves"."to" BETWEEN ? AND ?)
				OR
				(? BETWEEN "leaves"."from" AND "leaves"."to")
			)
		`, leave.EmployeeID, leave.From, leave.To, leave.From, leave.To, leave.From).
		Where(`
		(
			(
				"leaves"."approved_by_manager" IS NULL
				AND 
				"leaves"."approved_by_hr" IS NULL
				AND
				"leaves"."closed_automatically" IS NULL
			)
			OR
			(
				"leaves"."approved_by_manager" = ?
				AND
				"leaves"."approved_by_hr" IS NULL
				AND
				"leaves"."closed_automatically" IS NULL
			)
			OR
			(
				"leaves"."approved_by_manager" = ?
				AND
				"leaves"."approved_by_hr" = ?
			)
		)
		`, &truee, &truee, &truee).
		Count(&count).Error; err != nil {
		return false, nil
	}

	return count == 0, nil
}

func (repo *leaveRepo) CreateLeave(ctx context.Context, leave entity.Leave) error {
	tx := repo.db.WithContext(ctx).Begin()

	if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&entity.Leave{}).Create(&leave).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := 0; i < len(leave.Childs)+1; i++ {
		if i == 0 {
			sql := repo.generateUpdateLeaveQuotaSql(leave.Type, false)
			if sql == "" {
				continue
			}
			if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.From, leave.To), leave.EmployeeID).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			sql := repo.generateUpdateLeaveQuotaSql(leave.Childs[i-1].Type, false)
			if sql == "" {
				continue
			}
			if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.Childs[i-1].From, leave.Childs[i-1].To), leave.EmployeeID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

/*
*********************************
ACTOR: MANAGER
*********************************
*/
func (repo *leaveRepo) GetIncomingLeaveProposalForManager(ctx context.Context, managerId string, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()

	var leaves []entity.Leave
	var count int64

	// Only query the parent
	t := repo.db.WithContext(ctx).Model(&entity.Leave{}).
		Where(`"leaves"."manager_id" = ?`, managerId).
		Where("approved_by_manager IS NULL").
		Where("action_by_manager_at IS NULL").
		Where("approved_by_hr IS NULL").
		Where("action_by_hr_at IS NULL").
		Where("closed_automatically IS NULL").
		Where("parent_id IS NULL").
		Preload("Childs")

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "employees"."id" = "leaves"."employee_id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name)).Preload("Employee")
	} else {
		t = t.Preload("Employee")
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate)).Count(&count)
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."created_at") = ?`, tquery.Year).Count(&count)
	default:
		t = t.Count(&count)
	}

	t = t.Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset)

	if err := t.Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

func (repo *leaveRepo) SaveProcessedLeaveByManager(ctx context.Context, leave entity.Leave) error {
	tx := repo.db.WithContext(ctx).Begin()

	for i := 0; i < len(leave.Childs)+1; i++ {
		if i == 0 {
			// Process the parent leaves
			if err := tx.Exec("UPDATE leaves SET approved_by_manager = ?, action_by_manager_at = ?, rejection_reason = ? WHERE id = ?",
				leave.ApprovedByManager,
				leave.ActionByManagerAt,
				leave.RejectionReason,
				leave.ID).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Returns back the quota if the leave is rejected
			if !*leave.ApprovedByManager {
				sql := repo.generateUpdateLeaveQuotaSql(leave.Type, true)
				if sql == "" {
					continue
				}
				if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.From, leave.To), leave.EmployeeID).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		} else {
			// Process the child's leave
			if err := tx.Exec("UPDATE leaves SET approved_by_manager = ?, action_by_manager_at = ?, rejection_reason = ? WHERE id = ?",
				leave.Childs[i-1].ApprovedByManager,
				leave.Childs[i-1].ActionByManagerAt,
				leave.Childs[i-1].RejectionReason,
				leave.Childs[i-1].ID).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Returns back the quota if the child leave is rejected
			if !*leave.Childs[i-1].ApprovedByManager {
				sql := repo.generateUpdateLeaveQuotaSql(leave.Childs[i-1].Type, true)
				if sql == "" {
					continue
				}
				if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.Childs[i-1].From, leave.Childs[i-1].To), leave.EmployeeID).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (repo *leaveRepo) GetLeaveProposalHistoryForManager(ctx context.Context, managerId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()
	status := strings.ToLower(q.Status)

	var leaves []entity.Leave
	var count int64

	t := repo.db.WithContext(ctx).
		Model(&entity.Leave{}).
		Where(`"leaves"."manager_id" = ?`, managerId).
		Where("parent_id IS NULL")

	switch status {
	case "pending":
		t = t.Where("approved_by_manager IS TRUE AND approved_by_hr IS NULL AND closed_automatically IS NULL")
	case "approved":
		t = t.Where("approved_by_manager IS TRUE AND approved_by_hr IS TRUE AND closed_automatically IS NULL")
	case "rejected":
		t = t.Where("approved_by_manager IS NOT NULL AND (approved_by_manager IS FALSE OR approved_by_hr IS FALSE) AND closed_automatically IS NULL")
	case "closed":
		t = t.Where("closed_automatically IS TRUE")
	default:
		t = t.Where("approved_by_manager IS NOT NULL OR closed_automatically IS NOT NULL")
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "employees"."id" = "leaves"."employee_id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if err := t.Preload("Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

/*
*********************************
ACTOR: HR
*********************************
*/
func (repo *leaveRepo) GetIncomingLeaveProposalForHr(ctx context.Context, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()

	var leaves []entity.Leave
	var count int64
	truee := true

	// Only query the parent
	t := repo.db.WithContext(ctx).Model(&entity.Leave{}).
		Where("action_by_manager_at IS NOT NULL").
		Where("approved_by_manager = ?", &truee).
		Where("approved_by_hr IS NULL").
		Where("action_by_hr_at IS NULL").
		Where("closed_automatically IS NULL").
		Where("parent_id IS NULL").
		Preload("Childs")

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "employees"."id" = "leaves"."employee_id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name)).Preload("Employee")
	} else {
		t = t.Preload("Employee")
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."created_at") = ?`, tquery.Year)
	}

	if err := t.
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

func (repo *leaveRepo) SaveProcessedLeaveByHr(ctx context.Context, leave entity.Leave) error {
	tx := repo.db.WithContext(ctx).Begin()

	for i := 0; i < len(leave.Childs)+1; i++ {
		if i == 0 {
			// Process the parent leaves
			if err := tx.Exec("UPDATE leaves SET approved_by_hr = ?, action_by_hr_at = ?, rejection_reason = ?, hr_id = ? WHERE id = ?",
				leave.ApprovedByHr,
				leave.ActionByHrAt,
				leave.RejectionReason,
				leave.HrID,
				leave.ID).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Returns back the quota if the leave is rejected
			if !*leave.ApprovedByHr {
				sql := repo.generateUpdateLeaveQuotaSql(leave.Type, true)
				if sql == "" {
					continue
				}
				if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.From, leave.To), leave.EmployeeID).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		} else {
			// NOTE: Some child leave may already have been rejected by manager
			// Process the child's leave
			if *leave.Childs[i-1].ApprovedByManager {
				if err := tx.Exec("UPDATE leaves SET approved_by_hr = ?, action_by_hr_at = ?, rejection_reason = ?, hr_id = ? WHERE id = ?",
					leave.Childs[i-1].ApprovedByHr,
					leave.Childs[i-1].ActionByHrAt,
					leave.Childs[i-1].RejectionReason,
					leave.HrID,
					leave.Childs[i-1].ID).Error; err != nil {
					tx.Rollback()
					return err
				}

				// Returns back the quota if the child leave is rejected
				if !*leave.Childs[i-1].ApprovedByHr {
					sql := repo.generateUpdateLeaveQuotaSql(leave.Childs[i-1].Type, true)
					if sql == "" {
						continue
					}
					if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(leave.Childs[i-1].From, leave.Childs[i-1].To), leave.EmployeeID).Error; err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (repo *leaveRepo) GetLeaveProposalHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()
	tquery, _ := q.TimeQuery.Extract()
	status := strings.ToLower(q.Status)

	var leaves []entity.Leave
	var count int64

	t := repo.db.WithContext(ctx).
		Model(&entity.Leave{}).
		Where("parent_id IS NULL")

	switch status {
	case "pending":
		t = t.Where("(approved_by_manager IS TRUE OR approved_by_manager IS NULL) AND approved_by_hr IS NULL AND closed_automatically IS NULL")
	case "approved":
		t = t.Where("approved_by_manager IS TRUE AND approved_by_hr IS TRUE AND closed_automatically IS NULL")
	case "rejected":
		t = t.Where("(approved_by_manager IS FALSE OR approved_by_hr IS FALSE) AND closed_automatically IS NULL")
	case "closed":
		t = t.Where("closed_automatically IS TRUE")
	}

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."created_at" BETWEEN ? AND ?`, tquery.StartDate, tquery.EndDate))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."created_at") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."created_at") = ?`, tquery.Year)
	}

	if q.Name != "" {
		t = t.Joins(`INNER JOIN "employees" ON "employees"."id" = "leaves"."employee_id" AND "employees"."full_name" ILIKE ?`, utils.ToPatternMatching(q.Name))
	}

	if err := t.Preload("Employee").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

/*
*********************************
ACTOR: ALL
*********************************
*/
func (repo *leaveRepo) GetLeaveById(ctx context.Context, id string) (entity.Leave, error) {
	var leave entity.Leave

	if err := repo.db.WithContext(ctx).
		Model(&leave).
		Preload("Childs").
		Preload("Parent").
		// TODO: Delete .Manager here
		Preload("Employee.Manager").
		Preload("Manager").
		Preload("Hr").
		First(&leave, "id = ?", id).
		Error; err != nil {
		return leave, err
	}

	return leave, nil
}

func (repo *leaveRepo) WhosTakingLeave(ctx context.Context, q vo.CommonQuery) (vo.WhosTakingLeaveList, error) {
	tquery, _ := q.TimeQuery.Extract()

	startOfTheMonth := utils.GetStartOfTheMonthFromMonthAndYear(tquery.Month, tquery.Year)
	endOfTheMonth := utils.GetEndOfTheMonthFromMonthAndYear(tquery.Month, tquery.Year)

	var res vo.WhosTakingLeaveList = make(vo.WhosTakingLeaveList)

	rows, err := repo.db.WithContext(ctx).Raw(`
	SELECT d.leave_date, STRING_AGG(DISTINCT l.id::text, ',') AS leave_ids
	FROM (
		SELECT GENERATE_SERIES("from", "to", '1 day'::interval) AS leave_date
		FROM leaves
		WHERE
		approved_by_manager IS TRUE
		AND
		approved_by_hr IS TRUE
		AND
		(
			"from" BETWEEN ? AND ?
			OR
			"to" BETWEEN ? AND ?
		)
	) AS d
	JOIN leaves AS l ON d.leave_date BETWEEN l."from" AND l."to"
	WHERE EXTRACT(DOW FROM d.leave_date) NOT IN (0, 6)
	AND d.leave_date BETWEEN ? AND ?
	GROUP BY d.leave_date
	ORDER BY d.leave_date
	`, startOfTheMonth.In(utils.CURRENT_LOC), endOfTheMonth.In(utils.CURRENT_LOC), startOfTheMonth.In(utils.CURRENT_LOC), endOfTheMonth.In(utils.CURRENT_LOC), startOfTheMonth, endOfTheMonth).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t time.Time
		var ids string

		rows.Scan(&t, &ids)
		t = t.In(utils.CURRENT_LOC)

		var elements []vo.WhosTakingLeaveElements
		for _, v := range strings.Split(ids, ",") {
			var leave entity.Leave

			if err := repo.db.WithContext(ctx).
				Model(&leave).
				Preload("Employee.Role").
				Take(&leave, "id = ?", v).Error; err != nil {
				return nil, err
			}

			if leave.ApprovedByHr != nil && leave.ApprovedByManager != nil {
				if *leave.ApprovedByHr && *leave.ApprovedByManager {
					elements = append(elements, vo.WhosTakingLeaveElements{
						ID:       v,
						Avatar:   leave.Employee.Avatar,
						FullName: leave.Employee.FullName,
						Role:     leave.Employee.Role.Code,
						Type:     leave.Type.String(),
					})
				}
			}

		}

		if t.Day() < 10 {
			res["0"+strconv.Itoa(t.Day())] = elements
		} else {
			res[strconv.Itoa(t.Day())] = elements
		}
	}

	return res, nil
}

func (repo *leaveRepo) WhosTakingLeaveMobile(ctx context.Context, q vo.CommonQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	tquery, _ := q.TimeQuery.Extract()
	pquery := q.Pagination.MustExtract()

	startOfTheMonth := utils.GetStartOfTheMonthFromMonthAndYear(tquery.Month, tquery.Year)
	endOfTheMonth := utils.GetEndOfTheMonthFromMonthAndYear(tquery.Month, tquery.Year)

	var leaves []entity.Leave
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Leave{}).Where("approved_by_manager IS TRUE").Where("approved_by_hr IS TRUE")

	switch tquery.Option {
	case 1:
		t = t.Where(gorm.Expr(`"leaves"."from" BETWEEN ? AND ?`, startOfTheMonth, endOfTheMonth))
	case 2:
		t = t.Where(`EXTRACT(MONTH FROM "leaves"."from") = ?`, tquery.Month).Where(`EXTRACT(YEAR FROM "leaves"."from") = ?`, tquery.Year)
	}

	if err := t.Preload("Employee.Role").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&leaves).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return leaves, pquery.Compress(count), nil
}

/*
*************************************************
UTILS
*************************************************
*/
// generateUpdateLeaveQuotaSql generates an sql that will be used
// to update the employee's leave quota depending on leave type
// provided and whether it is reversed or not.
func (repo *leaveRepo) generateUpdateLeaveQuotaSql(leaveType entity.LeaveType, reverse bool) string {
	mapper := map[entity.LeaveType]string{
		entity.ANNUAL:   "yearly_count",
		entity.MARRIAGE: "marriage_count",
		entity.UNPAID:   "unpaid_count",
	}

	switch leaveType {
	case entity.UNPAID:
		if reverse {
			return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s - ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
		}
		return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s + ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
	case entity.ANNUAL, entity.MARRIAGE:
		if reverse {
			return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s + ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
		}
		return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s - ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
	default:
		return ""
	}
}
