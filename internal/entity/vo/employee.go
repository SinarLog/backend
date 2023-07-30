package vo

import "sinarlog.com/internal/entity"

type UpdateEmployeeData struct {
	ContractType *entity.ContractType `json:"contractType,omitempty"`
	JobId        string               `json:"jobId,omitempty"`
	RoleId       string               `json:"roleId,omitempty"`
	ManagerId    string               `json:"managerId,omitempty"`
	Status       *entity.Status       `json:"status,omitempty"`
}

type UpdateMyData struct {
	Id          string             `json:"id,omitempty"`
	PhoneNumber string             `json:"phoneNumber,omitempty"`
	Address     string             `json:"address,omitempty"`
	Contacts    []EmergencyContact `json:"contacts,omitempty"`
}

type UpdatePassword struct {
	NewPassword     string `json:"newPassword,omitempty"`
	ConfirmPassword string `json:"confirmPassword,omitempty"`
}

type EmergencyContact struct {
	Id          string          `json:"id,omitempty"`
	FullName    string          `json:"fullName,omitempty"`
	PhoneNumber string          `json:"phoneNumber,omitempty"`
	Relation    entity.Relation `json:"relation,omitempty"`
}
