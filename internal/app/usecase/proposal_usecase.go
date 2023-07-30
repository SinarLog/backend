package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

/*
*********************************
ACTOR: MANAGER
*********************************
*/
func (uc *leaveUseCase) SeeIncomingLeaveProposalsForManager(ctx context.Context, manager entity.Employee, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	// Set the order to request_date descendingly
	q.Pagination.Order = "created_at"
	q.Pagination.Sort = "DESC"
	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Query", err)
	}

	leaves, page, err := uc.leaveRepo.GetIncomingLeaveProposalForManager(ctx, manager.Id, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

func (uc *leaveUseCase) TakeActionOnLeaveProposalForManager(ctx context.Context, manager entity.Employee, action vo.LeaveAction) error {
	leave, err := uc.leaveRepo.GetLeaveById(ctx, action.Id)
	if err != nil {
		return NewRepositoryError("Leave", err)
	}

	// Validate that not the child leave is retrieve
	if leave.ParentID != nil {
		return NewDomainError("Leave", fmt.Errorf("processing only the leave overflow is not allowed"))
	}

	// Cheks if the leave was closed automatically
	if leave.ClosedAutomatically != nil {
		if *leave.ClosedAutomatically {
			return NewDomainError("Leave", fmt.Errorf("unable to process this leave because it has already been closed"))
		}
	}

	// Checks if this leave has been processed
	if leave.ApprovedByManager != nil || leave.ApprovedByHr != nil || leave.ActionByHrAt != nil || leave.ActionByManagerAt != nil {
		return NewDomainError("Leave", fmt.Errorf("unable to process your action because this leave has been proceesed"))
	}

	// Validate entity
	if err := action.Validate(); err != nil {
		return NewDomainError("Leave", err)
	}

	// Checks whether the leave has a child
	if leave.Childs != nil {
		// Validate the length of leave child and payload child
		if len(leave.Childs) != len(action.Childs) {
			return NewDomainError("Leave", fmt.Errorf("each leave overflows must have an action provided. Please include all overflows into action"))
		}

		// Validate that its true each of
		// the childs belongs to the parent
		var childIds []any
		for _, v := range leave.Childs {
			childIds = append(childIds, v.Id)
		}
		for _, v := range action.Childs {
			if err := validation.Validate(v.Id, validation.In(childIds...)); err != nil {
				return NewDomainError("Leave", fmt.Errorf("an overflow leave id does not belong to the associated leave"))
			}
		}
	}

	// Validate that its true the leave is directed for the current user.
	if leave.ManagerID == nil {
		return NewDomainError("Leave", fmt.Errorf("this leave request is not assigned for you"))
	} else if *leave.ManagerID != manager.Id {
		return NewDomainError("Leave", fmt.Errorf("you are not allowed to process this leave request"))
	}

	// NOTES: If the parent leave is rejected, all the children
	// will be rejected. On the other hand, if parent is approved
	// its children may be individual approved or rejected.
	now := time.Now().In(utils.CURRENT_LOC)
	shouldSendNotif := false

	// Check the action on the parent leave
	if !action.Approved {
		// Validate reason
		if err := validation.Validate(&action.Reason, validation.Required, validation.Length(20, 1000)); err != nil {
			return NewDomainError("Leave", fmt.Errorf("a rejected leave request must be provided with a reason"))
		}

		shouldSendNotif = true
		// Automatically rejects all leaves
		leave.ActionByManagerAt = &now
		leave.ApprovedByManager = &action.Approved
		leave.RejectionReason = action.Reason
		for i := 0; i < len(leave.Childs); i++ {
			leave.Childs[i].ActionByManagerAt = &now
			leave.Childs[i].ApprovedByManager = &action.Approved
			leave.Childs[i].RejectionReason = "This leave is rejected because the associated leave was rejected. All overflows will be automatically rejected if the main leave is rejected"
		}
	} else {
		leave.ActionByManagerAt = &now
		leave.ApprovedByManager = &action.Approved
		for i := 0; i < len(leave.Childs); i++ {
			leave.Childs[i].ApprovedByManager = &action.Childs[i].Approved
			leave.Childs[i].ActionByManagerAt = &now
			if !action.Childs[i].Approved {
				shouldSendNotif = true
				// Validate reason
				if err := validation.Validate(&action.Childs[i].Reason, validation.Required, validation.Length(20, 1000)); err != nil {
					return NewDomainError("Leave", fmt.Errorf("a rejected leave request must be provided with a reason"))
				}
				leave.Childs[i].RejectionReason = action.Childs[i].Reason
			}
		}
	}

	if err := uc.leaveRepo.SaveProcessedLeaveByManager(ctx, leave); err != nil {
		return NewRepositoryError("Leave", err)
	}

	if shouldSendNotif {
		go uc.sendProcessedLeaveProposalByManager(leave)
	}

	return nil
}

func (uc *leaveUseCase) RetrieveLeaveProposalsHistoryForManager(ctx context.Context, manager entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "created_at"

	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

	leaves, page, err := uc.leaveRepo.GetLeaveProposalHistoryForManager(ctx, manager.Id, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

/*
*********************************
ACTOR: HR
*********************************
*/
func (uc *leaveUseCase) SeeIncomingLeaveProposalsForHr(ctx context.Context, q vo.IncomingLeaveProposals) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	// Set the order to request_date descendingly
	q.Pagination.Order = "created_at"
	q.Pagination.Sort = "DESC"
	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Query", err)
	}

	leaves, page, err := uc.leaveRepo.GetIncomingLeaveProposalForHr(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

func (uc *leaveUseCase) TakeActionOnLeaveProposalForHr(ctx context.Context, hr entity.Employee, action vo.LeaveAction) error {
	leave, err := uc.leaveRepo.GetLeaveById(ctx, action.Id)
	if err != nil {
		return NewRepositoryError("Leave", err)
	}

	// Validate that not the child leave is retrieve
	if leave.ParentID != nil {
		return NewDomainError("Leave", fmt.Errorf("processing only the leave overflow is not allowed"))
	}

	// Cheks if the leave was closed automatically
	if leave.ClosedAutomatically != nil {
		if *leave.ClosedAutomatically {
			return NewDomainError("Leave", fmt.Errorf("unable to process this leave because it has already been closed"))
		}
	}

	// Check whether this leave has been processed by manager and is approved
	if leave.ApprovedByManager == nil {
		return NewDomainError("Leave", fmt.Errorf("this leave has not been process by manager"))
	} else if !*leave.ApprovedByManager {
		return NewDomainError("Leave", fmt.Errorf("this leave has been rejected by manager and no action needed from your side"))
	}

	// Checks if the leave has been processed
	if leave.ApprovedByHr != nil || leave.ActionByHrAt != nil {
		return NewDomainError("Leave", fmt.Errorf("unable to process your action because this leave has been processed"))
	}

	// Validate entity
	if err := action.Validate(); err != nil {
		return NewDomainError("Leave", err)
	}

	// Checks if the leave has a child
	if leave.Childs != nil {
		// Validate that the payload also has actions for the children
		if action.Childs == nil {
			return NewDomainError("Leave", fmt.Errorf("the overflows of this leave must also be processed"))
		}

		// Variable to store the count of the required action childs
		var requiredAction int
		// Variable to store the required child ids
		var childsRequiredIds []any

		for _, v := range leave.Childs {
			// Validate whether all the childs has been processed too
			if v.ApprovedByManager == nil {
				return NewDomainError("Leave", fmt.Errorf("there exists an overflow associated with the leave that has not been processed"))
			} else if *v.ApprovedByManager {
				requiredAction++
				childsRequiredIds = append(childsRequiredIds, v.Id)
			}
		}

		// Validate whether the required action sent is the same with the domain
		if requiredAction != len(action.Childs) {
			return NewDomainError("Leave", fmt.Errorf("all unfinished process of this leave overflows must be taken into action"))
		}

		// Validate that its true each of the childs belongs to the parent
		for _, v := range action.Childs {
			if err := validation.Validate(v.Id, validation.In(childsRequiredIds...)); err != nil {
				return NewDomainError("Leave", fmt.Errorf("an overflow leave id does not belong to the associated leave"))
			}
		}
	}

	// NOTES: If the parent leave is rejected, all the children
	// will be rejected. On the other hand, if parent is approved
	// its children may be individual approved or rejected.
	now := time.Now().In(utils.CURRENT_LOC)
	leave.HrID = &hr.Id

	if !action.Approved {
		if err := validation.Validate(&action.Reason, validation.Required, validation.Length(20, 1000)); err != nil {
			return NewDomainError("Leave", fmt.Errorf("a rejected leave requested must be provided with a well constructed reason"))
		}
		leave.ActionByHrAt = &now
		leave.ApprovedByHr = &action.Approved
		leave.RejectionReason = action.Reason
		for i := 0; i < len(leave.Childs); i++ {
			if *leave.Childs[i].ApprovedByManager {
				leave.Childs[i].ActionByHrAt = &now
				leave.Childs[i].ApprovedByHr = &action.Approved
				leave.Childs[i].RejectionReason = "This leave is rejected because the associated leave was rejected. All overflows will be automatically rejected if the main leave is rejected"
			}
		}
	} else {
		leave.ActionByHrAt = &now
		leave.ApprovedByHr = &action.Approved
		for i := 0; i < len(leave.Childs); i++ {
			if *leave.Childs[i].ApprovedByManager {
				// Find the approriate action
				var k vo.LeaveAction
				for _, v := range action.Childs {
					if v.Id == leave.Childs[i].Id {
						k = v
					}
				}

				leave.Childs[i].ApprovedByHr = &k.Approved
				leave.Childs[i].ActionByHrAt = &now
				if !k.Approved {
					if err := validation.Validate(&k.Reason, validation.Required, validation.Length(20, 1000)); err != nil {
						return NewDomainError("Leave", fmt.Errorf("a rejected leave request must be provided with a well constructed reason"))
					}
					leave.Childs[i].RejectionReason = k.Reason
				}
			}
		}
	}

	if err := uc.leaveRepo.SaveProcessedLeaveByHr(ctx, leave); err != nil {
		return NewRepositoryError("Leave", err)
	}

	go uc.sendProcessedLeaveProposalByHr(leave)

	return nil
}

func (uc *leaveUseCase) RetrieveLeaveProposalsHistoryForHr(ctx context.Context, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "created_at"

	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

	leaves, page, err := uc.leaveRepo.GetLeaveProposalHistoryForHr(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

/*
*************************************************
MAILER HELPERS
*************************************************
*/
func (uc *leaveUseCase) sendProcessedLeaveProposalByManager(leave entity.Leave) {
	data := make(map[string]any)
	data["RequesteeName"] = leave.Employee.FullName
	data["LeaveType"] = strings.ToLower(leave.Type.String())
	data["Reason"] = leave.RejectionReason
	data["Approved"] = *leave.ApprovedByManager
	if utils.CountNumberOfDays(leave.From, leave.To) <= 1 {
		data["At"] = leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
	} else {
		data["From"] = leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
		data["To"] = leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
	}
	if len(leave.Childs) != 0 {
		data["HaveAdditionals"] = true
		childs := make([]map[string]any, len(leave.Childs))
		for _, v := range leave.Childs {
			item := make(map[string]any)
			r := []rune(strings.ToLower(v.Type.String()))
			item["LeaveType"] = string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
			if utils.CountNumberOfDays(v.From, v.To) <= 1 {
				item["At"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
			} else {
				item["From"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
				item["To"] = v.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
			}
			item["Approved"] = *v.ApprovedByManager
			item["Reason"] = v.RejectionReason
			childs = append(childs, item)
		}
		data["Additionals"] = childs
	} else {
		data["HaveAdditionals"] = false
	}

	if err := uc.mailService.SendEmail(leave.Employee.Email, service.PROCESSED_LEAVE_BY_MANAGER, data); err != nil {
		log.Printf("unable to send mail to %s due to %s\n", leave.Employee.Email, err.Error())
	}
}

func (uc *leaveUseCase) sendProcessedLeaveProposalByHr(leave entity.Leave) {
	data := make(map[string]any)
	data["RequesteeName"] = leave.Employee.FullName
	data["LeaveType"] = strings.ToLower(leave.Type.String())
	data["Reason"] = leave.RejectionReason
	data["Approved"] = *leave.ApprovedByHr
	if utils.CountNumberOfDays(leave.From, leave.To) <= 1 {
		data["At"] = leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
	} else {
		data["From"] = leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
		data["To"] = leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
	}
	if len(leave.Childs) != 0 {
		data["HaveAdditionals"] = true
		childs := make([]map[string]any, len(leave.Childs))
		for _, v := range leave.Childs {
			item := make(map[string]any)
			r := []rune(strings.ToLower(v.Type.String()))
			item["LeaveType"] = string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
			if utils.CountNumberOfDays(v.From, v.To) <= 1 {
				item["At"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
			} else {
				item["From"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
				item["To"] = v.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
			}
			if v.ApprovedByHr != nil {
				item["Approved"] = *v.ApprovedByHr
			} else {
				item["Approved"] = *v.ApprovedByManager
			}
			item["Reason"] = v.RejectionReason
			childs = append(childs, item)
		}
		data["Additionals"] = childs
	} else {
		data["HaveAdditionals"] = false
	}

	if err := uc.mailService.SendEmail(leave.Employee.Email, service.PROCESSED_LEAVE_BY_HR, data); err != nil {
		log.Printf("unable to send mail to %s due to %s\n", leave.Employee.Email, err.Error())
	}
}
