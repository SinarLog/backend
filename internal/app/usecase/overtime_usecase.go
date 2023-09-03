package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

/*
*********************************
ACTOR: STAFF
*********************************
*/
func (uc *attendanceUseCase) RetrieveMyOvertimeSubmissions(ctx context.Context, employee entity.Employee, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	q.Pagination.Sort = "DESC"
	q.Pagination.Order = "created_at"

	overtimes, page, err := uc.attRepo.GetMyOvertimeSubmissions(ctx, employee.ID, q)
	if err != nil {
		return nil, page, NewRepositoryError("Overtime", err)
	}

	return overtimes, page, nil
}

/*
*********************************
ACTOR: MANAGER
*********************************
*/
func (uc *attendanceUseCase) SeeIncomingOvertimeSubmissionsForManager(ctx context.Context, manager entity.Employee, q vo.IncomingOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	q.Pagination.Sort = "DESC"
	q.Pagination.Order = "created_at"

	overtimes, page, err := uc.attRepo.GetIncomingOvertimeSubmissionsForManager(ctx, manager.ID, q)
	if err != nil {
		return nil, page, NewRepositoryError("Overtime", err)
	}
	return overtimes, page, nil
}

func (uc *attendanceUseCase) RetrieveOvertimeSubmission(ctx context.Context, overtimeId string) (entity.Overtime, error) {
	overtime, err := uc.attRepo.GetOvertimeById(ctx, overtimeId)
	if err != nil {
		return entity.Overtime{}, NewRepositoryError("Overtime", err)
	}

	return overtime, nil
}

func (uc *attendanceUseCase) TakeActionOnOvertimeSubmissionByManager(ctx context.Context, manager entity.Employee, action vo.OvertimeSubmissionAction) error {
	overtime, err := uc.attRepo.GetOvertimeById(ctx, action.ID)
	if err != nil {
		NewRepositoryError("Overtime", err)
	}

	// Checks if the overtime has been processed
	if overtime.ApprovedByManager != nil && overtime.ActionByManagerAt != nil {
		return NewDomainError("Overtime", fmt.Errorf("this overtime submission has been processed"))
	}

	// Checks if the overtime has been closed automatically
	if overtime.ClosedAutomatically != nil {
		if *overtime.ClosedAutomatically {
			return NewDomainError("Overtime", fmt.Errorf("unable to process a closed overtime submission"))
		}
	}

	// Check if the overtime is intended for the current user
	if *overtime.ManagerID != manager.ID {
		return NewDomainError("Overtime", fmt.Errorf("you are not allowed to process this overtime submission"))
	}

	// Start processing the request
	now := time.Now().In(utils.CURRENT_LOC)
	overtime.ApprovedByManager = &action.Approved
	overtime.ActionByManagerAt = &now
	overtime.RejectionReason = action.Reason

	if err := uc.attRepo.SaveProcessedOvertimeSubmissionByManager(ctx, overtime); err != nil {
		return NewRepositoryError("Overtime", err)
	}

	go uc.sendProcessedOvertimeSubmissionByManagerEmail(overtime)

	return nil
}

func (uc *attendanceUseCase) RetrieveOvertimeSubmissionsHistoryForManager(ctx context.Context, manager entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "created_at"

	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Overtime", err)
	}

	overtimes, page, err := uc.attRepo.GetOvertimeSubmissionHistoryForManager(ctx, manager.ID, q)
	if err != nil {
		return nil, page, NewRepositoryError("Overtime", err)
	}

	return overtimes, page, nil
}

func (uc *attendanceUseCase) RetrieveMyEmployeesOvertime(ctx context.Context, manager entity.Employee, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	q.Pagination.Sort = "DESC"
	q.Pagination.Order = "created_at"

	// Verify that the employee is indeed under the manager
	employee, err := uc.emplRepo.GetEmployeeById(ctx, employeeId)
	if err != nil {
		return nil, vo.PaginationDTOResponse{}, NewRepositoryError("Leave", err)
	}
	if employee.ManagerID == nil {
		return nil, vo.PaginationDTOResponse{}, NewDomainError("Employee", fmt.Errorf("the employee you're requesting to view is not your staff"))
	}
	if *employee.ManagerID != manager.ID {
		return nil, vo.PaginationDTOResponse{}, NewDomainError("Employee", fmt.Errorf("the employee you're requesting to view is not your staff"))
	}

	overtimes, page, err := uc.attRepo.GetMyOvertimeSubmissions(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Overtime", err)
	}

	return overtimes, page, nil
}

/*
*********************************
ACTOR: HR
*********************************
*/
func (uc *attendanceUseCase) RetrieveOvertimeSubmissionsHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "created_at"

	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Overtime", err)
	}

	overtimes, page, err := uc.attRepo.GetOvertimeSubmissionHistoryForHr(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Overtime", err)
	}

	return overtimes, page, nil
}

func (uc *attendanceUseCase) RetrieveAnEmployeeOvertimes(ctx context.Context, employeeId string, q vo.MyOvertimeSubmissionsQuery) ([]entity.Overtime, vo.PaginationDTOResponse, error) {
	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

	overtimes, page, err := uc.attRepo.GetMyOvertimeSubmissions(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return overtimes, page, nil
}

/*
*************************************************
MAILER HELPERS
*************************************************
*/
func (uc *attendanceUseCase) sendProcessedOvertimeSubmissionByManagerEmail(overtime entity.Overtime) {
	data := map[string]any{
		"RequesteeName":   overtime.Attendance.Employee.FullName,
		"Approved":        *overtime.ApprovedByManager,
		"RejectionReason": overtime.RejectionReason,
	}

	if err := uc.mailService.SendEmail(overtime.Attendance.Employee.Email, service.PROCESSED_OVERTIME_SUBMISSION, data); err != nil {
		log.Printf("Unable to send processed overtime submission email due to: %s", err.Error())
	}
}
