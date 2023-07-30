package repo

import (
	"context"

	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type employeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *employeeRepo {
	return &employeeRepo{db}
}

func (repo *employeeRepo) CreateNewEmployee(ctx context.Context, employee entity.Employee) error {
	if err := repo.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Model(&employee).
		Create(&employee).
		Error; err != nil {
		return err
	}

	return nil
}

func (repo *employeeRepo) GetEmployeeFullNameById(ctx context.Context, id string) (name string, err error) {
	if err = repo.db.
		WithContext(ctx).
		Raw("SELECT full_name FROM employees WHERE id = ?", id).
		Scan(&name).
		Error; err != nil {
		return name, err
	}

	return name, nil
}

// GetAllManagersList retrieve the managers brief profile
func (repo *employeeRepo) GetAllManagersList(ctx context.Context) ([]entity.Employee, error) {
	var managers []entity.Employee

	if err := repo.db.WithContext(ctx).
		Model(&entity.Employee{}).
		InnerJoins("Role", repo.db.Where(&entity.Role{
			Code: "mngr",
		})).
		Find(&managers).Error; err != nil {
		return nil, err
	}

	return managers, nil
}

// GetEmployeeBiodataById retrieves all of the employees information.
// This preloads all of the employee's associations regarding its biodata.
func (repo *employeeRepo) GetEmployeeFullProfileById(ctx context.Context, id string) (entity.Employee, error) {
	var employee entity.Employee

	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Preload("Job").
		Preload("Role").
		Preload("Manager").
		Preload("EmployeeBiodata").
		Preload("EmployeeLeavesQuota").
		Preload("EmployeesEmergencyContacts").
		First(&employee, "id = ?", id).Error; err != nil {
		return employee, err
	}

	return employee, nil
}

// GetEmployeeById retrieve employee's brief profile by id.
func (repo *employeeRepo) GetEmployeeById(ctx context.Context, id string) (entity.Employee, error) {
	var employee entity.Employee

	if err := repo.db.WithContext(ctx).Model(&employee).First(&employee, "id = ?", id).Error; err != nil {
		return employee, err
	}
	return employee, nil
}

func (repo *employeeRepo) GetEmployeeSimpleInformationById(ctx context.Context, id string) (entity.Employee, error) {
	var employee entity.Employee

	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Preload("Role").
		Preload("Job").
		Preload("Manager").
		Preload("EmployeeBiodata").
		First(&employee, "id = ?", id).Error; err != nil {
		return employee, err
	}

	return employee, nil
}

func (repo *employeeRepo) GetEmployeeChangesLog(ctx context.Context, employeeId string, q vo.CommonQuery) ([]entity.EmployeeDataHistoryLog, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()

	var changes []entity.EmployeeDataHistoryLog
	var count int64

	if err := repo.db.WithContext(ctx).
		Model(&entity.EmployeeDataHistoryLog{}).
		Where("employee_id = ?", employeeId).
		Preload("UpdatedBy.Job").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&changes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return changes, pquery.Compress(count), nil
}

func (repo *employeeRepo) SetEmployeeStatusTo(ctx context.Context, employeeId string, status entity.Status) error {
	if err := repo.db.
		WithContext(ctx).
		Exec("UPDATE employees SET status = ? WHERE id = ?",
			status,
			employeeId,
		).
		Error; err != nil {
		return err
	}

	return nil
}

func (repo *employeeRepo) GetLeaveQuotaByEmployeeId(ctx context.Context, employeeId string) (entity.EmployeeLeavesQuota, error) {
	var quota entity.EmployeeLeavesQuota

	if err := repo.db.WithContext(ctx).
		Model(&entity.EmployeeLeavesQuota{}).
		First(&quota, "employee_id = ?", employeeId).
		Error; err != nil {
		return quota, err
	}

	return quota, nil
}

func (repo *employeeRepo) GetBiodataByEmployeeId(ctx context.Context, id string) (entity.EmployeeBiodata, error) {
	var biodata entity.EmployeeBiodata

	if err := repo.db.WithContext(ctx).
		Model(&entity.EmployeeBiodata{}).
		First(&biodata, "employee_id = ?", id).
		Error; err != nil {
		return biodata, err
	}

	return biodata, nil
}

func (repo *employeeRepo) GetAllEmployees(ctx context.Context, employeeId, role string, q vo.AllEmployeeQuery) ([]entity.Employee, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()

	var employees []entity.Employee
	var count int64

	t := repo.db.WithContext(ctx).Model(&entity.Employee{}).Preload("Job")

	switch role {
	case "hr":
		t = t.Joins(`INNER JOIN "roles" ON "roles"."id" = "employees"."role_id" AND ("employees"."id" = ? OR "roles"."code" <> ?)`, employeeId, "hr")
	case "mngr", "staff":
		t = t.Joins(`INNER JOIN "roles" ON "roles"."id" = "employees"."role_id" AND "roles"."code" <> ?`, "hr").
			Where("resigned_at IS NULL").
			Where("resigned_by_id IS NULL")
	}

	if q.JobId != "" {
		t = t.Where("job_id = ?", q.JobId)
	}

	if q.FullName != "" {
		t = t.Where("full_name ILIKE ?", utils.ToPatternMatching(q.FullName))
	}

	if err := t.
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&employees).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return employees, pquery.Compress(count), nil
}

func (repo *employeeRepo) UpdateEmployeeWorkInfo(ctx context.Context, employee entity.Employee) error {
	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Omit("EmployeeBiodata").
		Omit("EmployeesEmergencyContacts").
		Omit("EmployeeLeavesQuota").
		Save(&employee).Error; err != nil {
		return err
	}

	return nil
}

func (repo *employeeRepo) UpdatePersonalData(ctx context.Context, employee entity.Employee, log entity.EmployeeDataHistoryLog) error {
	tx := repo.db.WithContext(ctx).Begin()

	if err := tx.Model(&employee.EmployeeBiodata).Updates(entity.EmployeeBiodata{
		PhoneNumber: employee.EmployeeBiodata.PhoneNumber,
		Address:     employee.EmployeeBiodata.Address,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&entity.EmployeesEmergencyContact{}).Save(&employee.EmployeesEmergencyContacts).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&log).Save(&log).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (repo *employeeRepo) UpdateAvatar(ctx context.Context, employee entity.Employee) error {
	if err := repo.db.WithContext(ctx).
		Model(&employee).
		Where("id = ?", employee.Id).
		Update("avatar", employee.Avatar).Error; err != nil {
		return err
	}

	return nil
}
