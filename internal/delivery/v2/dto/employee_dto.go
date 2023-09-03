package dto

import (
	"mime/multipart"

	"sinarlog.com/internal/entity"
)

type CreateNewEmployeeRequest struct {
	FullName             string                `form:"fullName" binding:"required"`
	Email                string                `form:"email" binding:"required"`
	ContractType         entity.ContractType   `form:"contractType" binding:"required"`
	Avatar               *multipart.FileHeader `form:"avatar"`
	NIK                  string                `form:"nik" binding:"required"`
	NPWP                 string                `form:"npwp" binding:"required"`
	Gender               entity.Gender         `form:"gender" binding:"required"`
	Religion             entity.Religion       `form:"religion" binding:"required"`
	PhoneNumber          string                `form:"phoneNumber" binding:"required"`
	Address              string                `form:"address" binding:"required"`
	BirthDate            string                `form:"birthDate" binding:"required"`
	MaritalStatus        bool                  `form:"maritalStatus"`
	EmergencyFullName    string                `form:"emergencyFullName" binding:"required"`
	EmergencyPhoneNumber string                `form:"emergencyPhoneNumber" binding:"required"`
	EmergencyRelation    entity.Relation       `form:"emergencyRelation" binding:"required"`

	RoleID    string `form:"roleId" binding:"required"`
	JobID     string `form:"jobId" binding:"required"`
	ManagerID string `form:"managerId"`
}

type BriefEmployeeListResponse struct {
	ID       string `json:"id,omitempty"`
	FullName string `json:"fullName,omitempty"`
	Status   string `json:"status,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Job      string `json:"job,omitempty"`
	JoinDate string `json:"joinDate,omitempty"`
}

type EmployeeBiodataResponse struct {
	EmployeeID    string `json:"employeeId,omitempty"`
	NIK           string `json:"nik,omitempty"`
	NPWP          string `json:"npwp,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Religion      string `json:"religion,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
	Address       string `json:"address,omitempty"`
	BirthDate     string `json:"birthDate,omitempty"`
	MaritalStatus bool   `json:"maritalStatus"`
}

type EmployeeLeaveQuotaResponse struct {
	EmployeeID    string `json:"employeeId,omitempty"`
	YearlyCount   int    `json:"yearlyCount"`
	UnpaidCount   int    `json:"unpaidCount"`
	MarriageCount int    `json:"marriageCount"`
}

type EmployeeEmergencyContactResponse struct {
	ID          string `json:"id"`
	EmployeeID  string `json:"employeeId,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	Relation    string `json:"relation,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type EmployeeChangesLogs struct {
	ID        string                    `json:"id,omitempty"`
	UpdatedBy BriefEmployeeListResponse `json:"updatedBy,omitempty"`
	Changes   map[string]any            `json:"changes,omitempty"`
	UpdatedAt string                    `json:"updatedAt,omitempty"`
}

type EmployeeFullProfileResponse struct {
	ID           string  `json:"id,omitempty"`
	FullName     string  `json:"fullName,omitempty"`
	Email        string  `json:"email,omitempty"`
	ContractType string  `json:"contractType,omitempty"`
	Avatar       string  `json:"avatar,omitempty"`
	Status       string  `json:"status,omitempty"`
	JoinDate     string  `json:"joinDate,omitempty"`
	ResignDate   *string `json:"resignDate,omitempty"`

	Biodata           EmployeeBiodataResponse            `json:"biodata,omitempty"`
	EmergencyContacts []EmployeeEmergencyContactResponse `json:"emergencyContacts,omitempty"`
	LeaveQuota        EmployeeLeaveQuotaResponse         `json:"leaveQuota,omitempty"`
	Logs              []EmployeeChangesLogs              `json:"logs,omitempty"`

	ManagerID *string                    `json:"managerId,omitempty"`
	Manager   *BriefEmployeeListResponse `json:"manager,omitempty"`
	// CreatedByID  *string `gorm:"type:uuid;default:null"`
	// CreatedBy    *Employee
	// ResignedByID *string `gorm:"type:uuid;default:null"`
	// ResignedBy   *Employee
	Role RoleResponse `json:"role,omitempty"`
	Job  JobResponse  `json:"job,omitempty"`
}

type UpdateAvatar struct {
	Avatar *multipart.FileHeader `form:"avatar"`
}
