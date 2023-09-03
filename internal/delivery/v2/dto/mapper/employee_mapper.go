package mapper

import (
	"fmt"
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

/*
*************************************************
REQUEST TO ENTITIES
*************************************************
*/

// Maps form fields to employee entities on create new employee
func MapCreateNewEmployeeRequestToEmployeeEntity(req dto.CreateNewEmployeeRequest) (entity.Employee, error) {
	res := entity.Employee{
		FullName:     req.FullName,
		Email:        req.Email,
		ContractType: req.ContractType,
		IsNewUser:    true,
		EmployeeBiodata: entity.EmployeeBiodata{
			NIK:           req.NIK,
			NPWP:          req.NPWP,
			Gender:        req.Gender,
			Religion:      req.Religion,
			PhoneNumber:   req.PhoneNumber,
			Address:       req.Address,
			MaritalStatus: req.MaritalStatus,
		},
		EmployeesEmergencyContacts: []entity.EmployeesEmergencyContact{
			{
				FullName:    req.EmergencyFullName,
				Relation:    req.EmergencyRelation,
				PhoneNumber: req.EmergencyPhoneNumber,
			},
		},
		RoleID: req.RoleID,
		JobID:  req.JobID,
	}

	birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
	if err != nil {
		return entity.Employee{}, fmt.Errorf("birth date format must be yyyy-mm-dd: %w", err)
	}

	res.EmployeeBiodata.BirthDate = birthDate

	if req.ManagerID != "" {
		res.ManagerID = &req.ManagerID
	}

	return res, nil
}

/*
*************************************************
ENTITIES TO RESPONSE
*************************************************
*/
func MapEmployeeListToBriefEmployeeListResponse(employees []entity.Employee) []dto.BriefEmployeeListResponse {
	var res []dto.BriefEmployeeListResponse

	for _, v := range employees {
		res = append(res, dto.BriefEmployeeListResponse{
			ID:       v.ID,
			FullName: v.FullName,
			Status:   string(v.Status),
			Email:    v.Email,
			Avatar:   v.Avatar,
			JoinDate: v.JoinDate.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Job:      v.Job.Name,
		})
	}

	return res
}

func MapManagersListToResponse(managers []entity.Employee) []dto.BriefEmployeeListResponse {
	var res []dto.BriefEmployeeListResponse

	for _, v := range managers {
		res = append(res, dto.BriefEmployeeListResponse{
			ID:       v.ID,
			FullName: v.FullName,
			Email:    v.Email,
		})
	}

	return res
}

func MapEmployeeBiodataToResponse(biodata entity.EmployeeBiodata) dto.EmployeeBiodataResponse {
	return dto.EmployeeBiodataResponse{
		EmployeeID:    biodata.EmployeeID,
		NIK:           biodata.NIK,
		NPWP:          biodata.NPWP,
		Gender:        string(biodata.Gender),
		Religion:      string(biodata.Religion),
		PhoneNumber:   biodata.PhoneNumber,
		Address:       biodata.Address,
		BirthDate:     biodata.BirthDate.In(utils.CURRENT_LOC).Format(time.DateOnly),
		MaritalStatus: biodata.MaritalStatus,
	}
}

func MapEmployeeLeaveQuotaToResponse(quota entity.EmployeeLeavesQuota) dto.EmployeeLeaveQuotaResponse {
	return dto.EmployeeLeaveQuotaResponse{
		EmployeeID:    quota.EmployeeID,
		YearlyCount:   quota.YearlyCount,
		UnpaidCount:   quota.UnpaidCount,
		MarriageCount: quota.MarriageCount,
	}
}

func MapEmployeeEmergencyContactToResponse(contact entity.EmployeesEmergencyContact) dto.EmployeeEmergencyContactResponse {
	return dto.EmployeeEmergencyContactResponse{
		ID:          contact.ID,
		EmployeeID:  contact.EmployeeID,
		FullName:    contact.FullName,
		Relation:    string(contact.Relation),
		PhoneNumber: contact.PhoneNumber,
	}
}

func MapEmployeeChangesLogToResponse(logs []entity.EmployeeDataHistoryLog) []dto.EmployeeChangesLogs {
	var res []dto.EmployeeChangesLogs

	for _, v := range logs {
		res = append(res, dto.EmployeeChangesLogs{
			ID: v.ID,
			UpdatedBy: dto.BriefEmployeeListResponse{
				ID:       v.UpdatedByID,
				FullName: v.UpdatedBy.FullName,
				Email:    v.UpdatedBy.Email,
				Avatar:   v.UpdatedBy.Avatar,
				Job:      v.UpdatedBy.Job.Name,
			},
			Changes:   v.Changes,
			UpdatedAt: v.UpdatedAt.In(utils.CURRENT_LOC).Format(time.RFC1123),
		})
	}

	return res
}

func MapEmployeeFullProfileToResponse(employee entity.Employee) dto.EmployeeFullProfileResponse {
	res := dto.EmployeeFullProfileResponse{
		ID:           employee.ID,
		FullName:     employee.FullName,
		Email:        employee.Email,
		ContractType: string(employee.ContractType),
		Avatar:       employee.Avatar,
		Status:       string(employee.Status),
		JoinDate:     employee.JoinDate.In(utils.CURRENT_LOC).Format(time.DateOnly),
		Biodata:      MapEmployeeBiodataToResponse(employee.EmployeeBiodata),
		LeaveQuota:   MapEmployeeLeaveQuotaToResponse(employee.EmployeeLeavesQuota),
		Role: dto.RoleResponse{
			ID:   employee.RoleID,
			Name: employee.Role.Name,
			Code: employee.Role.Code,
		},
		Job: dto.JobResponse{
			ID:   employee.JobID,
			Name: employee.Job.Name,
		},
	}

	if employee.ResignedAt != nil {
		resignedDate := employee.ResignedAt.In(utils.CURRENT_LOC).Format(time.DateOnly)
		res.ResignDate = &resignedDate
	}

	if employee.ManagerID != nil && employee.Manager != nil {
		manager := dto.BriefEmployeeListResponse{
			ID:       *employee.ManagerID,
			FullName: employee.Manager.FullName,
			Status:   string(employee.Manager.Status),
			Email:    employee.Manager.Email,
			Avatar:   employee.Manager.Avatar,
		}
		res.Manager = &manager
	}

	for _, v := range employee.EmployeesEmergencyContacts {
		res.EmergencyContacts = append(res.EmergencyContacts, MapEmployeeEmergencyContactToResponse(v))
	}

	return res
}
