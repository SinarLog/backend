package usecase

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type leaveUseCase struct {
	leaveRepo    repo.ILeaveRepo
	emplRepo     repo.IEmployeeRepo
	configRepo   repo.IConfigRepo
	mailService  service.IMailerService
	notifService service.INotifService
	bktService   service.IBucketService
}

func NewLeaveUseCase(
	leaveRepo repo.ILeaveRepo,
	emplRepo repo.IEmployeeRepo,
	configRepo repo.IConfigRepo,
	mailService service.IMailerService,
	notifService service.INotifService,
	bktService service.IBucketService,
) *leaveUseCase {
	return &leaveUseCase{
		leaveRepo:    leaveRepo,
		emplRepo:     emplRepo,
		configRepo:   configRepo,
		mailService:  mailService,
		notifService: notifService,
		bktService:   bktService,
	}
}

/*
*********************************
ACTOR: STAFF and MANAGER
*********************************
*/
// RetrieveMyLeaves retrieves the current user's
// leave request. All types of leave request whether
// pending or not, will be retrieved.
func (uc *leaveUseCase) RetrieveMyLeaves(ctx context.Context, employee entity.Employee, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	leaves, page, err := uc.leaveRepo.GetMyLeaveRequestsList(ctx, employee.ID, q)
	if err != nil {
		return nil, vo.PaginationDTOResponse{}, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

// RetrieveLeaveRequest retrieve a leave by ID. It is
// used to see detail of the leave and/or taking action
// to the leave.
func (uc *leaveUseCase) RetrieveLeaveRequest(ctx context.Context, id string) (entity.Leave, error) {
	leave, err := uc.leaveRepo.GetLeaveById(ctx, id)
	if err != nil {
		return leave, NewRepositoryError("Leave", err)
	}

	return leave, nil
}

// RetrieveMyQuotas retrieves leave quotas for the
// requested employee ID. This will be mark deprecated
// and be changed for RetrieveMyLeaveAggregate to show
// leave aggregates in the dashboard.
func (uc *leaveUseCase) RetrieveMyQuotas(ctx context.Context, employee entity.Employee) (entity.EmployeeLeavesQuota, error) {
	quota, err := uc.emplRepo.GetLeaveQuotaByEmployeeId(ctx, employee.ID)
	if err != nil {
		return quota, NewRepositoryError("Employee", err)
	}

	return quota, nil
}

// RequestLeave requests for a leave. It does neccessary
// checks to a leave request, like whether the employee has
// sufficient quota depending on the leave type requested.
// It then returns an error or a report regarding the leave
// request.
func (uc *leaveUseCase) RequestLeave(ctx context.Context, employee entity.Employee, leave entity.Leave) (entity.LeaveReport, error) {
	leave.EmployeeID = employee.ID

	// Check availability
	isAvailable, err := uc.leaveRepo.CheckDateAvailability(ctx, leave)
	if err == nil {
		if !isAvailable {
			return entity.LeaveReport{}, NewDomainError("Leave", fmt.Errorf("there has been an overlap of dates in your leave requests"))
		}
	} else {
		return entity.LeaveReport{}, NewRepositoryError("Leave", err)
	}

	// Validate by leave domain
	if err := leave.Validate(); err != nil {
		return entity.LeaveReport{}, NewDomainError("Leave", err)
	}

	// Validate employee domain
	if err := employee.ValidateLeave(); err != nil {
		return entity.LeaveReport{}, NewDomainError("Employee", err)
	}

	return uc.createLeaveReport(ctx, employee, leave)
}

// ApplyForLeave calls RequestLeave to do the neccessary
// checking. Together with the payload from the requestee,
// in then do more validation to the leave request. Once
// validation has passed, it creates the neccessary records
// to the leave request, and sends email to the manager (if the
// requestee is not a manager).
func (uc *leaveUseCase) ApplyForLeave(ctx context.Context, employee entity.Employee, decision vo.UserLeaveDecision, attachment multipart.File) error {
	report, err := uc.RequestLeave(ctx, employee, decision.Parent)
	if err != nil {
		return err
	}

	// Start creating leave domain
	var parentsEndDate time.Time
	if report.IsLeaveLeakage {
		parentsEndDate = utils.AddNumOfWorkingDays(time.Date(
			decision.Parent.From.Year(),
			decision.Parent.From.Month(),
			decision.Parent.From.Day(),
			23, 59, 59, 0,
			utils.CURRENT_LOC,
		), report.RemainingQuotaForRequestedType-1)
	} else {
		parentsEndDate = utils.GetWorkingDay(time.Date(
			decision.Parent.To.Year(),
			decision.Parent.To.Month(),
			decision.Parent.To.Day(),
			23, 59, 59, 0,
			utils.CURRENT_LOC,
		))
	}

	// Create the parent leave.
	parent := entity.Leave{
		BaseModelID: entity.BaseModelID{ID: uuid.NewString()},
		EmployeeID:  employee.ID,
		Employee:    employee,
		From:        decision.Parent.From,
		To:          parentsEndDate,
		Type:        decision.Parent.Type,
		Reason:      decision.Parent.Reason,
	}

	// Checks whether the requestee is a manager
	if employee.ManagerID == nil {
		now := time.Now().In(utils.CURRENT_LOC)
		truee := true
		parent.ApprovedByManager = &truee
		parent.ActionByManagerAt = &now
	} else {
		parent.ManagerID = employee.ManagerID
	}

	// Checks whether there has been a leave leakage
	if report.IsLeaveLeakage {
		if err := decision.ValidateExcessSumOfDays(report); err != nil {
			return NewDomainError("Leave", err)
		}

		// This will be used for validating the leave excess types
		var types []any
		for _, v := range report.AvailableExcessTypes {
			types = append(types, any(v))
		}

		nextChildsStartDate := utils.GetWorkingDay(time.Date(
			parentsEndDate.Year(),
			parentsEndDate.Month(),
			parentsEndDate.Day()+1, 0, 0, 0, 0,
			utils.CURRENT_LOC,
		))

		// See user decision
		for i, v := range decision.Overflows {
			if v.Count != 0 {
				// Validate leave excess type
				if err := validation.Validate(&v.Type, validation.Required, validation.In(types...)); err != nil {
					return NewDomainError("Leave", fmt.Errorf("the excess type of %s is not available in the option", strings.ToLower(v.Type.String())))
				}
				// Validate leave excess quota
				if v.Count > report.AvailableExcessQuotas[i] {
					return NewDomainError("Leave", fmt.Errorf("the amount of excess for %s exceeded limit", v.Type))
				}

				parent.Childs = append(parent.Childs, entity.Leave{
					EmployeeID: parent.EmployeeID,
					From:       nextChildsStartDate,
					To:         utils.AddNumOfWorkingDays(time.Date(nextChildsStartDate.Year(), nextChildsStartDate.Month(), nextChildsStartDate.Day(), 23, 59, 59, 0, utils.CURRENT_LOC), v.Count-1),
					Type:       v.Type,
					Reason: fmt.Sprintf("This is an extension leave of %s made by %s from %s to %s. This reason is autogenerated.",
						parent.Type,
						utils.GetFirstNameFromFullName(employee.FullName),
						parent.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
						parent.To.In(utils.CURRENT_LOC).Format(time.DateOnly)),
					ParentID:          &parent.ID,
					ManagerID:         parent.ManagerID,
					ApprovedByManager: parent.ApprovedByManager,
					ActionByManagerAt: parent.ActionByManagerAt,
				})
				nextChildsStartDate = utils.AddNumOfWorkingDays(time.Date(nextChildsStartDate.Year(), nextChildsStartDate.Month(), nextChildsStartDate.Day(), 0, 0, 0, 0, utils.CURRENT_LOC), v.Count)
			}
		}
	}

	// If an attachment is provided, upload to the bucket
	if attachment != nil {
		url, err := uc.bktService.CreateLeaveAttachment(ctx, parent.ID, attachment)
		if err != nil {
			return NewServiceError("Bucket", err)
		}
		parent.AttachmentUrl = url
	}

	if err := uc.leaveRepo.CreateLeave(ctx, parent); err != nil {
		uc.bktService.DeleteLeaveAttachment(ctx, parent.ID)
		return NewRepositoryError("Leave", err)
	}

	// Preparing to send notification only if requestee is a staff
	if employee.ManagerID != nil {
		manager, err := uc.emplRepo.GetEmployeeById(ctx, *employee.ManagerID)
		if err != nil {
			return NewErrorWithReport(
				"Leave",
				500,
				ErrUnexpected,
				fmt.Errorf("unable to retrieve manager's record"),
				"Your leave attendance has been successfully saved. However, we were unable to send your manager a notification. Please let your manager know directly of your leave request.",
			)
		}

		// Sending notification by redis pubsub or email
		receivers, err := uc.notifService.SendLeaveRequestNotification(ctx, manager, employee)
		if err != nil || receivers == 0 {
			go uc.sendLeaveRequestToManager(manager, employee, parent)
		}
	}

	return nil
}

/*
*********************************
ACTOR: MANAGER
*********************************
*/
func (uc *leaveUseCase) RetrieveMyEmployeesLeaveHistory(ctx context.Context, manager entity.Employee, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	if _, err := q.CommonQuery.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

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

	leaves, page, err := uc.leaveRepo.GetMyLeaveRequestsList(ctx, employeeId, q)
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
func (uc *leaveUseCase) RetrieveAnEmployeeLeaves(ctx context.Context, employeeId string, q vo.LeaveProposalHistoryQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

	leaves, page, err := uc.leaveRepo.GetMyLeaveRequestsList(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return leaves, page, nil
}

/*
*********************************
ACTOR: ALL
*********************************
*/
func (uc *leaveUseCase) RetrieveWhosTakingLeave(ctx context.Context, q vo.CommonQuery) (vo.WhosTakingLeaveList, error) {
	calendars, err := uc.leaveRepo.WhosTakingLeave(ctx, q)
	if err != nil {
		return nil, NewRepositoryError("Leave", err)
	}

	return calendars, nil
}

func (uc *leaveUseCase) RetrieveWhosTakingLeaveMobile(ctx context.Context, q vo.CommonQuery) ([]entity.Leave, vo.PaginationDTOResponse, error) {
	calendars, page, err := uc.leaveRepo.WhosTakingLeaveMobile(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return calendars, page, nil
}

/*
*************************************************
UTILS
*************************************************
*/

// createLeaveReport is a helper function to analyze a
// leave request. It does validation like checking quota,
// overlapping dates, and so on. If there are no errors
// provided, it returns a report showing whether the leave
// request can be proceed (without taking further steps)
// or a further step is required. Given a leave request,
// it could detect whether the duration of a leave request
// exceeded the limit quota based on the type. It then suggests
// an overflow to other leave types. Hence, user is needed
// to decide whether it wants to overflow to other types
// or change the duration.
func (uc *leaveUseCase) createLeaveReport(ctx context.Context, employee entity.Employee, leave entity.Leave) (entity.LeaveReport, error) {
	var report entity.LeaveReport

	// Query employee's leave quota
	quota, err := uc.emplRepo.GetLeaveQuotaByEmployeeId(ctx, employee.ID)
	if err != nil {
		return report, NewRepositoryError("Leave", err)
	}

	leaveDuration := utils.CountNumberOfWorkingDays(leave.From, leave.To)

	switch leave.Type {
	case entity.MARRIAGE:
		if quota.MarriageCount <= 0 {
			return report, NewDomainError("Leave", fmt.Errorf("there are no more quota for leave type %s. Please consider selecting other type", strings.ToLower(leave.Type.String())))
		}
		report.RequestType = entity.MARRIAGE
		report.RemainingQuotaForRequestedType = quota.MarriageCount

		config, err := uc.configRepo.GetConfiguration(ctx)
		if err != nil {
			return report, NewRepositoryError("Config", err)
		}
		if utils.CountNumberOfDays(time.Now().In(utils.CURRENT_LOC), leave.From) <= config.AcceptanceLeaveInterval {
			return report, NewDomainError("Leave", fmt.Errorf("a marriage leave request must be submitted %d days from now", config.AcceptanceLeaveInterval))
		}

		biodata, err := uc.emplRepo.GetBiodataByEmployeeId(ctx, employee.ID)
		if err != nil {
			return report, NewRepositoryError("Leave", err)
		}

		if biodata.MaritalStatus {
			return report, NewDomainError("Leave", fmt.Errorf("employee is already married"))
		}

		if leaveDuration > quota.MarriageCount {
			report.IsLeaveLeakage = true
			report.ExcessLeaveDuration = leaveDuration - quota.MarriageCount

			// Available excess from here should be annual and unpaid
			if quota.YearlyCount > 0 {
				report.AvailableExcessTypes = append(report.AvailableExcessTypes, entity.ANNUAL)
				report.AvailableExcessQuotas = append(report.AvailableExcessQuotas, quota.YearlyCount)
			}

			report.AvailableExcessTypes = append(report.AvailableExcessTypes, entity.UNPAID)
			report.AvailableExcessQuotas = append(report.AvailableExcessQuotas, 365)
		}

	case entity.ANNUAL:
		if quota.YearlyCount <= 0 {
			return report, NewDomainError("Leave", fmt.Errorf("there are no more quota for leave type %s. Please consider selecting other type", strings.ToLower(leave.Type.String())))
		}
		report.RequestType = entity.ANNUAL
		report.RemainingQuotaForRequestedType = quota.YearlyCount

		config, err := uc.configRepo.GetConfiguration(ctx)
		if err != nil {
			return report, NewRepositoryError("Config", err)
		}
		if utils.CountNumberOfDays(time.Now().In(utils.CURRENT_LOC), leave.From) < config.AcceptanceLeaveInterval {
			return report, NewDomainError("Leave", fmt.Errorf("an annual leave request must be submitted %d days from now", config.AcceptanceLeaveInterval))
		}

		if leaveDuration > quota.YearlyCount {
			report.IsLeaveLeakage = true
			report.ExcessLeaveDuration = leaveDuration - quota.YearlyCount

			// Available excess types from here are only unpaid...
			report.AvailableExcessTypes = append(report.AvailableExcessTypes, entity.UNPAID)
			report.AvailableExcessQuotas = append(report.AvailableExcessQuotas, 365)
		}

	case entity.UNPAID:
		report.RequestType = entity.UNPAID
		report.RemainingQuotaForRequestedType = 365

		if leaveDuration > report.RemainingQuotaForRequestedType {
			return report, NewDomainError("Leave", fmt.Errorf("maximum allowed unpaid leave duration is %d days", report.RemainingQuotaForRequestedType))
		}

		config, err := uc.configRepo.GetConfiguration(ctx)
		if err != nil {
			return report, NewRepositoryError("Config", err)
		}
		if utils.CountNumberOfDays(time.Now().In(utils.CURRENT_LOC), leave.From) <= config.AcceptanceLeaveInterval {
			return report, NewDomainError("Leave", fmt.Errorf("an unpaid leave request must be submitted %d days from now", config.AcceptanceLeaveInterval))
		}

	case entity.SICK:
		report.IsLeaveLeakage = false
		report.RequestType = entity.SICK
		report.RemainingQuotaForRequestedType = -1
	}

	return report, nil
}

/*
*************************************************
MAILER HELPERS
*************************************************
*/
func (uc *leaveUseCase) sendLeaveRequestToManager(manager, employee entity.Employee, parent entity.Leave) {
	data := make(map[string]any)
	data["RequesteeName"] = employee.FullName
	data["ManagerName"] = manager.FullName
	r := []rune(strings.ToLower(parent.Type.String()))
	data["LeaveType"] = string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
	data["Reason"] = parent.Reason
	if utils.CountNumberOfDays(parent.From, parent.To) <= 1 {
		data["At"] = parent.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
	} else {
		data["From"] = parent.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
		data["To"] = parent.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
	}
	if len(parent.Childs) != 0 {
		data["HaveAdditionals"] = true
		childs := make([]map[string]any, len(parent.Childs))
		for _, v := range parent.Childs {
			item := make(map[string]any)
			r := []rune(strings.ToLower(v.Type.String()))
			item["LeaveType"] = string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
			if utils.CountNumberOfDays(v.From, v.To) <= 1 {
				item["At"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
			} else {
				item["From"] = v.From.In(utils.CURRENT_LOC).Format(time.DateOnly)
				item["To"] = v.To.In(utils.CURRENT_LOC).Format(time.DateOnly)
			}
			childs = append(childs, item)
		}
		data["Additionals"] = childs
	} else {
		data["HaveAdditionals"] = false
	}

	if err := uc.mailService.SendEmail(manager.Email, service.FWD_LEAVE_PROPOSAL, data); err != nil {
		log.Printf("unable to send mail to %s due to %s\n", manager.Email, err.Error())
	}
}
