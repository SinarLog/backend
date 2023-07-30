package usecase

// This usecase is use for employee management... Such as creating a new employee,
// or see friends list, get employee detail  and so on

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type employeesUseCase struct {
	emplRepo    repo.IEmployeeRepo
	configRepo  repo.IConfigRepo
	sharedRepo  repo.ISharedRepo
	credRepo    repo.ICredentialRepo
	dkService   service.IDoorkeeperService
	mailService service.IMailerService
	bktService  service.IBucketService
}

func NewEmployeeUseCase(
	emplRepo repo.IEmployeeRepo,
	configRepo repo.IConfigRepo,
	sharedRepo repo.ISharedRepo,
	credRepo repo.ICredentialRepo,
	dkService service.IDoorkeeperService,
	mailService service.IMailerService,
	bktService service.IBucketService,
) *employeesUseCase {
	return &employeesUseCase{
		emplRepo:    emplRepo,
		configRepo:  configRepo,
		sharedRepo:  sharedRepo,
		credRepo:    credRepo,
		dkService:   dkService,
		mailService: mailService,
		bktService:  bktService,
	}
}

/*
*********************************
ACTOR: ALL
*********************************
*/
func (uc *employeesUseCase) RetrieveEmployeesList(ctx context.Context, requestee entity.Employee, q vo.AllEmployeeQuery) ([]entity.Employee, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "join_date"
	role, err := uc.sharedRepo.GetRoleById(ctx, requestee.RoleID)
	if err != nil {
		return nil, vo.PaginationDTOResponse{}, NewRepositoryError("Role", err)
	}

	employees, page, err := uc.emplRepo.GetAllEmployees(ctx, requestee.Id, role.Code, q)
	if err != nil {
		return employees, page, NewRepositoryError("Employee", err)
	}
	return employees, page, nil
}

func (uc *employeesUseCase) RetrieveMyProfile(ctx context.Context, user entity.Employee) (entity.Employee, error) {
	employee, err := uc.emplRepo.GetEmployeeFullProfileById(ctx, user.Id)
	if err != nil {
		return entity.Employee{}, NewRepositoryError("Employee", err)
	}

	return employee, nil
}

/*
***********************
* ACTOR: HR
***********************
 */
func (uc *employeesUseCase) RegisterNewEmployee(ctx context.Context, creator, payload entity.Employee, avatar multipart.File) error {
	payload.JoinDate = time.Now().In(utils.CURRENT_LOC)
	payload.IsNewUser = true
	payload.Status = entity.UNAVAILABLE
	payload.Id = uuid.NewString()
	payload.EmployeeLeavesQuota.EmployeeID = payload.Id

	if !payload.EmployeeBiodata.MaritalStatus {
		// Query config record
		config, err := uc.configRepo.GetConfiguration(ctx)
		if err != nil {
			return NewRepositoryError("Configurations", err)
		}
		payload.EmployeeLeavesQuota.MarriageCount = config.DefaultMarriageQuota
	}

	// Query Role
	role, err := uc.sharedRepo.GetRoleById(ctx, payload.RoleID)
	if err != nil {
		return NewNotFoundError("Role", err)
	}
	payload.Role = role

	// Query job
	job, err := uc.sharedRepo.GetJobById(ctx, payload.JobID)
	if err != nil {
		return NewNotFoundError("Job", err)
	}
	payload.Job = job

	// Set who created this employee record
	payload.CreatedById = &creator.Id
	payload.CreatedBy = &creator

	// Validate entity
	if err := payload.ValidateNewEmployee(); err != nil {
		return NewDomainError("Employee", err)
	}

	// Generate password
	generatedPassword := utils.GenerateRandomPassword()
	hashedPassword, err := uc.dkService.HashPassword(generatedPassword)
	if err != nil {
		return NewServiceError("Employee", err)
	}
	payload.Password = string(hashedPassword)

	// If avatar is provided, upload to the bucket
	if avatar != nil {
		url, err := uc.bktService.CreateAvatar(ctx, payload.Id, avatar)
		if err != nil {
			return NewServiceError("Bucket", err)
		}
		payload.Avatar = url
	}

	// Persist
	if err := uc.emplRepo.CreateNewEmployee(ctx, payload); err != nil {
		uc.bktService.DeleteAvatar(ctx, payload.Id)
		return NewRepositoryError("Employee", err)
	}

	// Prepare for sending email
	dataForMail := map[string]any{
		"FullName": payload.FullName,
		"Email":    payload.Email,
		"Password": generatedPassword,
	}
	if payload.ManagerID != nil {
		managerFullName, err := uc.emplRepo.GetEmployeeFullNameById(ctx, *payload.ManagerID)
		if err != nil {
			return NewRepositoryError("Employee", err)
		}
		dataForMail["ManagerFullName"] = managerFullName
		dataForMail["IsStaff"] = true
	}

	go uc.sendMailToNewEmployee(dataForMail)

	return nil
}

func (uc *employeesUseCase) ViewManagersList(ctx context.Context) ([]entity.Employee, error) {
	managers, err := uc.emplRepo.GetAllManagersList(ctx)
	if err != nil {
		return nil, NewRepositoryError("Employee", err)
	}

	return managers, nil
}

func (uc *employeesUseCase) RetrieveEmployeeFullProfile(ctx context.Context, requestee entity.Employee, employeeId string) (entity.Employee, error) {
	if requestee.ManagerID == nil {
		role, err := uc.sharedRepo.GetRoleById(ctx, requestee.RoleID)
		if err != nil {
			return entity.Employee{}, NewRepositoryError("Role", err)
		}

		switch role.Code {
		case "hr":
			employee, err := uc.emplRepo.GetEmployeeFullProfileById(ctx, employeeId)
			if err != nil {
				return entity.Employee{}, NewRepositoryError("Employee", err)
			}

			return employee, nil
		case "mngr":
			employee, err := uc.emplRepo.GetEmployeeById(ctx, employeeId)
			if err != nil {
				return entity.Employee{}, NewRepositoryError("Employee", err)
			}
			if employee.ManagerID == nil {
				employee, err = uc.emplRepo.GetEmployeeSimpleInformationById(ctx, employeeId)
				if err != nil {
					return employee, NewRepositoryError("Employee", err)
				}
				return employee, nil
			} else if *employee.ManagerID == requestee.Id {
				employee, err = uc.emplRepo.GetEmployeeFullProfileById(ctx, employeeId)
				if err != nil {
					return employee, NewRepositoryError("Employee", err)
				}
				return employee, nil
			} else {
				employee, err = uc.emplRepo.GetEmployeeSimpleInformationById(ctx, employeeId)
				if err != nil {
					return employee, NewRepositoryError("Employee", err)
				}
				return employee, nil
			}
		}
	}

	employee, err := uc.emplRepo.GetEmployeeSimpleInformationById(ctx, employeeId)
	if err != nil {
		return employee, NewRepositoryError("Employee", err)
	}
	return employee, nil
}

func (uc *employeesUseCase) RetrieveEmployeeBiodata(ctx context.Context, id string) (entity.EmployeeBiodata, error) {
	biodata, err := uc.emplRepo.GetBiodataByEmployeeId(ctx, id)
	if err != nil {
		return biodata, NewRepositoryError("Biodata", err)
	}

	return biodata, nil
}

func (uc *employeesUseCase) UpdateEmployeeData(ctx context.Context, hr entity.Employee, employeeId string, payload vo.UpdateEmployeeData) error {
	var manager entity.Employee
	var role entity.Role
	var job entity.Job
	var changes map[string]any = make(map[string]any)
	var logs entity.EmployeeDataHistoryLog

	// Query the employee
	employee, err := uc.emplRepo.GetEmployeeFullProfileById(ctx, employeeId)
	if err != nil {
		return NewRepositoryError("Employee", err)
	}

	logs.Employee = employee
	logs.EmployeeID = employeeId
	logs.UpdatedBy = hr
	logs.UpdatedByID = hr.Id

	// Update status
	if payload.Status != nil {
		if *payload.Status != employee.Status {
			change := make(map[string]string)
			change["prev"] = string(employee.Status)
			change["new"] = string(*payload.Status)
			changes["status"] = any(change)
			employee.Status = *payload.Status

			if *payload.Status == entity.RESIGNED {
				now := time.Now().In(utils.CURRENT_LOC)
				employee.ResignedAt = &now
				employee.ResignedById = &hr.Id
				employee.ResignedBy = &hr
			}
		}
	}

	// Update employment type
	if payload.ContractType != nil {
		if *payload.ContractType != employee.ContractType {
			change := make(map[string]string)
			change["prev"] = string(employee.ContractType)
			change["new"] = string(*payload.ContractType)
			changes["type"] = any(change)
			employee.ContractType = *payload.ContractType
		}
	}

	// Update role if given
	if payload.RoleId != "" {
		if payload.RoleId != employee.RoleID {
			role, err = uc.sharedRepo.GetRoleById(ctx, payload.RoleId)
			if err != nil {
				return NewRepositoryError("Role", err)
			}
			change := make(map[string]string)
			change["prev"] = employee.Role.Name
			change["new"] = role.Name
			changes["role"] = any(change)
			employee.Role = role
			employee.RoleID = payload.RoleId

			if employee.Role.Code != "staff" {
				employee.ManagerID = nil
				employee.Manager = nil
			}
		}
	}

	// Update job if given
	if payload.JobId != "" {
		if payload.JobId != employee.JobID {
			job, err = uc.sharedRepo.GetJobById(ctx, payload.JobId)
			if err != nil {
				return NewRepositoryError("Job", err)
			}
			change := make(map[string]string)
			change["prev"] = employee.Job.Name
			change["new"] = job.Name
			changes["job"] = any(change)
			employee.Job = job
			employee.JobID = payload.JobId
		}
	}

	// Update manager if given
	if payload.ManagerId != "" {
		// Employee previously has a manager
		if employee.ManagerID != nil {
			if payload.ManagerId != *employee.ManagerID {
				manager, err = uc.emplRepo.GetEmployeeById(ctx, payload.ManagerId)
				if err != nil {
					return NewRepositoryError("Employee", err)
				}
				change := make(map[string]string)
				change["prev"] = employee.Manager.FullName
				change["new"] = manager.FullName
				changes["manager"] = any(change)
				employee.Manager = &manager
				employee.ManagerID = &payload.ManagerId
			}
		} else {
			manager, err = uc.emplRepo.GetEmployeeById(ctx, payload.ManagerId)
			if err != nil {
				return NewRepositoryError("Employee", err)
			}
			change := make(map[string]string)
			change["prev"] = "No manager"
			change["new"] = manager.FullName
			changes["manager"] = any(change)
			employee.Manager = &manager
			employee.ManagerID = &payload.ManagerId
		}
	} else if payload.ManagerId == "" && employee.ManagerID != nil {
		change := make(map[string]string)
		change["prev"] = employee.Manager.FullName
		change["new"] = "No manager"
		changes["manager"] = any(change)
		employee.Manager = &manager
		employee.ManagerID = &payload.ManagerId
	}

	// Validate
	if err := employee.ValidateUpdateWorkInfo(); err != nil {
		return NewDomainError("Employee", err)
	}

	// Persist
	logs.Changes = changes
	employee.EmployeeDataHistoryLogs = append(employee.EmployeeDataHistoryLogs, logs)
	if err := uc.emplRepo.UpdateEmployeeWorkInfo(ctx, employee); err != nil {
		return NewRepositoryError("Employee", err)
	}

	return nil
}

func (uc *employeesUseCase) RetrieveEmployeeChangesLog(ctx context.Context, employeeId string, q vo.CommonQuery) ([]entity.EmployeeDataHistoryLog, vo.PaginationDTOResponse, error) {
	q.Pagination.Order = "updated_at"
	changes, page, err := uc.emplRepo.GetEmployeeChangesLog(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Employee", err)
	}

	return changes, page, nil
}

func (uc *employeesUseCase) UpdatePersonalData(ctx context.Context, user entity.Employee, payload vo.UpdateMyData) error {
	employee, err := uc.emplRepo.GetEmployeeFullProfileById(ctx, payload.Id)
	if err != nil {
		return NewRepositoryError("Employee", err)
	}

	var log entity.EmployeeDataHistoryLog
	var changes map[string]any = make(map[string]any)

	log.EmployeeID = user.Id
	log.Employee = user
	log.UpdatedByID = user.Id
	log.UpdatedBy = user

	if employee.EmployeeBiodata.Address != payload.Address {
		change := make(map[string]string)
		change["prev"] = employee.EmployeeBiodata.Address
		change["new"] = payload.Address
		changes["address"] = change
		employee.EmployeeBiodata.Address = payload.Address
	}

	if employee.EmployeeBiodata.PhoneNumber != payload.PhoneNumber {
		change := make(map[string]string)
		change["prev"] = employee.EmployeeBiodata.PhoneNumber
		change["new"] = payload.PhoneNumber
		changes["phone_number"] = change
		employee.EmployeeBiodata.PhoneNumber = payload.PhoneNumber
	}

	if len(employee.EmployeesEmergencyContacts) != len(payload.Contacts) {
		change := make(map[string]string)
		change["prev"] = fmt.Sprintf("%d emergency contacts", len(employee.EmployeesEmergencyContacts))
		change["new"] = fmt.Sprintf("%d emergency contacts", len(payload.Contacts))
		changes["emergency_contacts"] = change
		for _, v := range payload.Contacts {
			if v.Id == "" {
				employee.EmployeesEmergencyContacts = append(employee.EmployeesEmergencyContacts, entity.EmployeesEmergencyContact{
					EmployeeID:  employee.Id,
					Employee:    employee,
					FullName:    v.FullName,
					Relation:    v.Relation,
					PhoneNumber: v.PhoneNumber,
				})
			}
		}
	}

	log.Changes = changes
	if err := uc.emplRepo.UpdatePersonalData(ctx, employee, log); err != nil {
		return NewRepositoryError("Employee", err)
	}

	return nil
}

func (uc *employeesUseCase) UpdatePassword(ctx context.Context, employee entity.Employee, payload vo.UpdatePassword) error {
	if payload.NewPassword != payload.ConfirmPassword {
		return NewDomainError("Payload", fmt.Errorf("your new password and confirmation password is not the same"))
	}

	if !utils.IsStrongPassword(payload.NewPassword) {
		return NewDomainError("Employee", fmt.Errorf("password is not strong enough"))
	}

	hashedPassword, err := uc.dkService.HashPassword(payload.NewPassword)
	if err != nil {
		return NewServiceError("Employee", err)
	}

	employee.Password = string(hashedPassword)

	if err := uc.credRepo.UpdateEmployeePassword(ctx, employee); err != nil {
		return NewRepositoryError("Employee", err)
	}

	return nil
}

func (uc *employeesUseCase) UpdateProfilePic(ctx context.Context, employee entity.Employee, avatar multipart.File) error {
	if avatar == nil {
		if employee.Avatar == "" {
			return nil
		}
		if err := uc.bktService.DeleteAvatar(ctx, employee.Id); err != nil {
			return NewServiceError("Bucket", err)
		}
		employee.Avatar = ""
	} else {
		url, err := uc.bktService.CreateAvatar(ctx, employee.Id, avatar)
		if err != nil {
			return NewServiceError("Bucket", err)
		}
		employee.Avatar = url
	}

	if err := uc.emplRepo.UpdateAvatar(ctx, employee); err != nil {
		return NewRepositoryError("Employee", err)
	}

	return nil
}

/*
*************************************************
MAILER HELPERS
*************************************************
*/
func (uc *employeesUseCase) sendMailToNewEmployee(data map[string]any) {
	uc.mailService.SendEmail(data["Email"].(string), service.CRED, data)
}
