package repo

import (
	"context"

	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type ILeaveRepo interface {
	EmployeeIsOnLeaveToday(ctx context.Context, employeeId string) (bool, error)
	GetMyLeaveRequestsList(ctx context.Context, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)
	CheckDateAvailability(ctx context.Context, leave entity.Leave) (bool, error)
	CreateLeave(ctx context.Context, leave entity.Leave) error
	GetLeaveById(ctx context.Context, id string) (entity.Leave, error)

	GetIncomingLeaveProposalForManager(ctx context.Context, managerId string, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error)
	SaveProcessedLeaveByManager(ctx context.Context, leave entity.Leave) error
	GetLeaveProposalHistoryForManager(ctx context.Context, managerId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)

	GetIncomingLeaveProposalForHr(ctx context.Context, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error)
	SaveProcessedLeaveByHr(ctx context.Context, leave entity.Leave) error
	GetLeaveProposalHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)

	WhosTakingLeave(ctx context.Context, q vo.CommonQuery) (vo.WhosTakingLeaveList, error)
	WhosTakingLeaveMobile(ctx context.Context, q vo.CommonQuery) ([]entity.Leave, vo.PaginationDTOResponse, error)
}
