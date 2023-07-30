package repo

import (
	"context"
	"time"

	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type IAttendanceRepo interface {
	EmployeeHasClockedInToday(ctx context.Context, employeeId string) (bool, error)
	SaveClockInOTPTimestamp(ctx context.Context, emplooyeeId string, timestamp int64, exp time.Duration) error
	GetClockInOTPTimestamp(ctx context.Context, employeeId string) (int64, error)
	CreateNewAttendance(ctx context.Context, attendance entity.Attendance) error
	EmployeeHasActiveAttendance(ctx context.Context, employeeId string) (bool, error)
	GetTodaysAttendanceByEmployeeId(ctx context.Context, employeeId string) (entity.Attendance, error)
	GetActiveAttendanceByEmployeeId(ctx context.Context, employeeId string) (entity.Attendance, error)
	SumWeeklyOvertimeDurationByEmployeeId(ctx context.Context, employeeId string) (int, error)
	CloseAttendance(ctx context.Context, attendance entity.Attendance) error

	GetMyAttendancesHistory(ctx context.Context, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	GetStaffsAttendancesHistory(ctx context.Context, managerId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	GetEmployeesAttendanceHistory(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	GetEmployeesTodaysAttendances(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)

	GetIncomingOvertimeSubmissionsForManager(ctx context.Context, managerId string, q vo.IncomingOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	GetOvertimeById(ctx context.Context, overtimeId string) (entity.Overtime, error)
	SaveProcessedOvertimeSubmissionByManager(ctx context.Context, overtime entity.Overtime) error
	GetOvertimeSubmissionHistoryForManager(ctx context.Context, managerId string, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	GetOvertimeSubmissionHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)

	GetMyOvertimeSubmissions(ctx context.Context, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
}
