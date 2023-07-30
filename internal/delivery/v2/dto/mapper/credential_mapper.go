package mapper

import (
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

func MapLoginRequestToCredentialVO(req dto.LoginRequest) vo.Credential {
	return vo.Credential{
		Email:    req.Email,
		Password: req.Password,
	}
}

func MapToLoginResponse(employee entity.Employee, cred vo.Credential) dto.LoginResponse {
	return dto.LoginResponse{
		ID:        employee.Id,
		Email:     employee.Email,
		FullName:  employee.FullName,
		Avatar:    employee.Avatar,
		IsNewUser: employee.IsNewUser,
		Role: dto.RoleResponse{
			ID:   employee.RoleID,
			Name: employee.Role.Name,
			Code: employee.Role.Code,
		},
		Job: dto.JobResponse{
			ID:   employee.JobID,
			Name: employee.Job.Name,
		},
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
	}
}
