package repo

import (
	"context"

	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type IEmployeeRepo interface {
	CreateNewEmployee(ctx context.Context, employee entity.Employee) error
	UpdateEmployeeWorkInfo(ctx context.Context, employee entity.Employee) error
	UpdatePersonalData(ctx context.Context, employee entity.Employee, logs entity.EmployeeDataHistoryLog) error
	UpdateAvatar(ctx context.Context, employee entity.Employee) error

	GetEmployeeById(ctx context.Context, id string) (entity.Employee, error)
	GetBiodataByEmployeeId(ctx context.Context, employeeId string) (entity.EmployeeBiodata, error)
	GetEmployeeFullProfileById(ctx context.Context, id string) (entity.Employee, error)
	GetEmployeeFullNameById(ctx context.Context, id string) (string, error)
	GetLeaveQuotaByEmployeeId(ctx context.Context, employeeId string) (entity.EmployeeLeavesQuota, error)
	GetEmployeeSimpleInformationById(ctx context.Context, id string) (entity.Employee, error)

	GetEmployeeChangesLog(ctx context.Context, employeeId string, q vo.CommonQuery) ([]entity.EmployeeDataHistoryLog, vo.PaginationDTOResponse, error)

	SetEmployeeStatusTo(ctx context.Context, employeeId string, status entity.Status) error

	GetAllManagersList(ctx context.Context) ([]entity.Employee, error)
	GetAllEmployees(ctx context.Context, employeeId, role string, q vo.AllEmployeeQuery) ([]entity.Employee, vo.PaginationDTOResponse, error)
}
