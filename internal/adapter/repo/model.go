package repo

import "sinarlog.com/internal/entity"

func GetAllRelationalEntities() []any {
	return []any{
		&entity.Configuration{},
		&entity.Job{},
		&entity.Role{},
		&entity.Employee{},
		&entity.EmployeeBiodata{},
		&entity.EmployeesEmergencyContact{},
		&entity.EmployeeLeavesQuota{},
		&entity.EmployeeDataHistoryLog{},
		&entity.Attendance{},
		&entity.Leave{},
		&entity.Overtime{},
		&entity.ConfigurationChangesLog{},
	}
}
