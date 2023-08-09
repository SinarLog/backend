package usecase

import (
	"context"
	"mime/multipart"

	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type ICredentialUseCase interface {
	Login(ctx context.Context, cred vo.Credential) (entity.Employee, vo.Credential, error)
	Authorize(ctx context.Context, token string, roles ...any) (entity.Employee, error)
	ForgotPassword(ctx context.Context, email string) error
}

type IConfigUseCase interface {
	RetrieveConfiguration(ctx context.Context) (entity.Configuration, error)
	ChangeCompanyConfig(ctx context.Context, hr entity.Employee, payload entity.Configuration) error
	RetrieveChangesLogs(ctx context.Context, q vo.CommonQuery) ([]entity.ConfigurationChangesLog, vo.PaginationDTOResponse, error)
}

type IRoleUseCase interface {
	RetrieveRoles(ctx context.Context) ([]entity.Role, error)
}

type IJobUseCase interface {
	RetrieveJobs(ctx context.Context) ([]entity.Job, error)
}

type IEmployeeUseCase interface {
	RegisterNewEmployee(ctx context.Context, creator, payload entity.Employee, avatar multipart.File) error
	UpdateEmployeeData(ctx context.Context, hr entity.Employee, employeeId string, payload vo.UpdateEmployeeData) error
	UpdatePersonalData(ctx context.Context, user entity.Employee, payload vo.UpdateMyData) error
	UpdatePassword(ctx context.Context, employee entity.Employee, payload vo.UpdatePassword) error
	UpdateProfilePic(ctx context.Context, employee entity.Employee, avatar multipart.File) error

	RetrieveEmployeesList(ctx context.Context, requestee entity.Employee, query vo.AllEmployeeQuery) ([]entity.Employee, vo.PaginationDTOResponse, error)
	ViewManagersList(ctx context.Context) ([]entity.Employee, error)
	RetrieveEmployeeFullProfile(ctx context.Context, requestee entity.Employee, employeeId string) (entity.Employee, error)
	RetrieveMyProfile(ctx context.Context, user entity.Employee) (entity.Employee, error)
	RetrieveEmployeeBiodata(ctx context.Context, id string) (entity.EmployeeBiodata, error)
	RetrieveEmployeeChangesLog(ctx context.Context, employeeId string, q vo.CommonQuery) ([]entity.EmployeeDataHistoryLog, vo.PaginationDTOResponse, error)
}

type IAttendanceUseCase interface {
	RequestClockIn(ctx context.Context, employee entity.Employee) error
	ClockIn(ctx context.Context, employee entity.Employee, req vo.ClockInRequest) error
	RetrieveTodaysAttendance(ctx context.Context, employee entity.Employee) (entity.Attendance, error)
	RetrieveMyAttendanceHistory(ctx context.Context, employee entity.Employee, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	RequestClockOut(ctx context.Context, employee entity.Employee) (entity.OvertimeOnAttendanceReport, error)
	ClockOut(ctx context.Context, employee entity.Employee, payload vo.ClockOutPayload) error

	RetrieveOvertimeSubmission(ctx context.Context, overtimeId string) (entity.Overtime, error)
	SeeIncomingOvertimeSubmissionsForManager(ctx context.Context, manager entity.Employee, q vo.IncomingOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	TakeActionOnOvertimeSubmissionByManager(ctx context.Context, manager entity.Employee, action vo.OvertimeSubmissionAction) error
	RetrieveMyOvertimeSubmissions(ctx context.Context, employee entity.Employee, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	RetrieveOvertimeSubmissionsHistoryForManager(ctx context.Context, manager entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	RetrieveOvertimeSubmissionsHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)

	RetrieveMyStaffsAttendanceHistory(ctx context.Context, manager entity.Employee, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	RetrieveEmployeesAttendanceHistory(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	RetrieveEmployeesTodaysAttendances(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)

	RetrieveStaffAttendanceHistory(ctx context.Context, manager entity.Employee, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)
	RetrieveAnEmployeeAttendances(ctx context.Context, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error)

	RetrieveMyEmployeesOvertime(ctx context.Context, manager entity.Employee, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
	RetrieveAnEmployeeOvertimes(ctx context.Context, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error)
}

type ILeaveUseCase interface {
	RetrieveMyLeaves(ctx context.Context, employee entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)
	RetrieveMyQuotas(ctx context.Context, employee entity.Employee) (entity.EmployeeLeavesQuota, error)
	RequestLeave(ctx context.Context, employee entity.Employee, leave entity.Leave) (entity.LeaveReport, error)
	ApplyForLeave(ctx context.Context, employee entity.Employee, decision vo.UserLeaveDecision, attachment multipart.File) error
	RetrieveLeaveRequest(ctx context.Context, id string) (entity.Leave, error)

	SeeIncomingLeaveProposalsForManager(ctx context.Context, manager entity.Employee, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error)
	TakeActionOnLeaveProposalForManager(ctx context.Context, manager entity.Employee, action vo.LeaveAction) error
	RetrieveLeaveProposalsHistoryForManager(ctx context.Context, manager entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)

	SeeIncomingLeaveProposalsForHr(ctx context.Context, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error)
	TakeActionOnLeaveProposalForHr(ctx context.Context, hr entity.Employee, action vo.LeaveAction) error
	RetrieveLeaveProposalsHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)

	RetrieveWhosTakingLeave(ctx context.Context, q vo.CommonQuery) (vo.WhosTakingLeaveList, error)
	RetrieveWhosTakingLeaveMobile(ctx context.Context, q vo.CommonQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)

	RetrieveMyEmployeesLeaveHistory(ctx context.Context, manager entity.Employee, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)
	RetrieveAnEmployeeLeaves(ctx context.Context, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)
}

type IAnalyticsUseCase interface {
	RetrieveDashboardAnalyticsForEmployeeById(ctx context.Context, employeeId string) (vo.BriefLeaveAndAttendanceAnalytics, error)
	RetrieveDashboardAnalyticsHr(ctx context.Context) (vo.HrDashboardAnalytics, error)
}

type IChatUseCase interface {
	OpenChat(ctx context.Context, user entity.Employee, room entity.Room) (entity.Room, []entity.Chat, error)
	SendMessage(ctx context.Context, userId, roomId, message string) (entity.Chat, error)
	ListenMessage(ctx context.Context, userId, roomId string, channel chan entity.Chat) error
	DetachListener(ctx context.Context, userId, roomId string) error
}
