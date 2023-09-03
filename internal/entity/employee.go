package entity

import (
	"fmt"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"sinarlog.com/internal/utils"
)

type ContractType string

const (
	FULL_TIME ContractType = "FULL_TIME"
	CONTRACT  ContractType = "CONTRACT"
	INTERN    ContractType = "INTERN"
)

type Gender string

const (
	M Gender = "M"
	F Gender = "F"
)

type Religion string

const (
	CHRISTIAN Religion = "CHRISTIAN"
	MUSLIM    Religion = "MUSLIM"
	CATHOLIC  Religion = "CATHOLIC"
	BUDDHA    Religion = "BUDDHA"
	HINDU     Religion = "HINDU"
	CONFUCION Religion = "CONFUCION"
)

type Relation string

const (
	FATHER  Relation = "FATHER"
	MOTHER  Relation = "MOTHER"
	SIBLING Relation = "SIBLING"
	SPOUSE  Relation = "SPOUSE"
)

type Status string

const (
	AVAILABLE   Status = "AVAILABLE"
	UNAVAILABLE Status = "UNAVAILABLE"
	ON_LEAVE    Status = "ON_LEAVE"
	RESIGNED    Status = "RESIGNED"
)

type Employee struct {
	BaseModelID

	FullName     string       `gorm:"type:varchar(155)"`
	Email        string       `gorm:"type:varchar(150);index:,unique,type:btree"`
	Password     string       `gorm:"type:varchar(255)"`
	ContractType ContractType `gorm:"type:varchar(100)"`
	Avatar       string       `gorm:"type:varchar(255)"`
	Status       Status       `gorm:"type:varchar(100);default:'UNAVAILABLE'"`
	IsNewUser    bool
	JoinDate     time.Time

	EmployeeBiodata            EmployeeBiodata
	EmployeesEmergencyContacts []EmployeesEmergencyContact
	EmployeeLeavesQuota        EmployeeLeavesQuota
	EmployeeDataHistoryLogs    []EmployeeDataHistoryLog

	ManagerID    *string `gorm:"type:uuid"`
	Manager      *Employee
	CreatedByID  *string `gorm:"type:uuid;default:null"`
	CreatedBy    *Employee
	ResignedByID *string `gorm:"type:uuid;default:null"`
	ResignedBy   *Employee
	ResignedAt   *time.Time
	RoleID       string `gorm:"type:uuid"`
	Role         Role
	JobID        string `gorm:"type:uuid"`
	Job          Job

	BaseModelStamps
	BaseModelSoftDelete
}

type EmployeeBiodata struct {
	BaseModelID

	EmployeeID    string   `gorm:"type:uuid,uniqueIndex"`
	NIK           string   `gorm:"type:varchar(255);index:,unique,type:btree"`
	NPWP          string   `gorm:"type:varchar(255);index:,unique,type:btree"`
	Gender        Gender   `gorm:"type:varchar(10)"`
	Religion      Religion `gorm:"type:varchar(85)"`
	PhoneNumber   string   `gorm:"type:varchar(150);index:,unique,type:btree"`
	Address       string
	BirthDate     time.Time
	MaritalStatus bool

	BaseModelStamps
	BaseModelSoftDelete
}

type EmployeesEmergencyContact struct {
	BaseModelID

	EmployeeID  string `gorm:"type:uuid"`
	Employee    Employee
	FullName    string   `gorm:"type:varchar(255)"`
	Relation    Relation `goem:"type:varchar(150)"`
	PhoneNumber string   `gorm:"type:varchar(150)"`

	BaseModelStamps
	BaseModelSoftDelete
}

type EmployeeLeavesQuota struct {
	BaseModelID

	EmployeeID    string `gorm:"type:uuid"`
	YearlyCount   int    `gorm:"default:0"`
	UnpaidCount   int    `gorm:"default:0"`
	MarriageCount int    `gorm:"default:0"`

	BaseModelStamps
	BaseModelSoftDelete
}

type EmployeeDataHistoryLog struct {
	BaseModelID

	EmployeeID  string `gorm:"type:uuid"`
	Employee    Employee
	UpdatedByID string `gorm:"type:uuid"`
	UpdatedBy   Employee
	Changes     JSONB

	BaseModelStamps
	BaseModelSoftDelete
}

func (v EmployeeBiodata) Validate() error {
	now := time.Now().In(utils.CURRENT_LOC)

	return validation.ValidateStruct(&v,
		validation.Field(&v.NIK, validation.Required, validation.Match(regexp.MustCompile(
			`(1[1-9]|21|[37][1-6]|5[1-3]|6[1-5]|[89][12])\d{2}\d{2}([04][1-9]|[1256][0-9]|[37][01])(0[1-9]|1[0-2])\d{2}\d{4}$`,
		))),
		validation.Field(&v.NPWP, validation.Required, validation.Match(regexp.MustCompile(
			`[\d]{2}[.]([\d]{3})[.]([\d]{3})[.][\d][-]([\d]{3})[.]([\d]{3})$`,
		))),
		validation.Field(&v.Gender, validation.Required, validation.In(M, F)),
		validation.Field(&v.Religion, validation.Required, validation.In(CHRISTIAN, BUDDHA, MUSLIM, CONFUCION, CATHOLIC, HINDU)),
		validation.Field(&v.PhoneNumber, validation.Required, validation.Match(regexp.MustCompile(
			`^\+62-\d{3}-\d{4}-\d{4,5}$`),
		)),
		validation.Field(&v.Address, validation.Required, validation.Length(20, 1000)),
		validation.Field(&v.BirthDate, validation.Required, validation.Max(time.Date(now.Year()-18, now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), utils.CURRENT_LOC))),
	)
}

func (v EmployeesEmergencyContact) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.FullName, validation.Required, validation.Length(5, 100)),
		validation.Field(&v.Relation, validation.Required, validation.In(FATHER, MOTHER, SIBLING, SPOUSE)),
		validation.Field(&v.PhoneNumber, validation.Required, validation.Match(regexp.MustCompile(
			`^\+62-\d{3}-\d{4}-\d{4,5}$`),
		)),
	)
}

func (v Employee) ValidateNewEmployee() error {
	if err := validation.ValidateStruct(&v,
		validation.Field(&v.FullName, validation.Required, validation.Length(5, 100)),
		validation.Field(&v.ContractType, validation.Required, validation.In(FULL_TIME, CONTRACT, INTERN)),
		validation.Field(&v.JoinDate, validation.Required),
		validation.Field(&v.Avatar, validation.When(v.Avatar != "", validation.Required, is.URL)),
		validation.Field(&v.Status, validation.Required, validation.In(AVAILABLE, UNAVAILABLE, ON_LEAVE)),
		validation.Field(&v.ManagerID, validation.When(v.Role.Code == "staff",
			validation.Required,
			validation.By(func(value interface{}) error {
				v, ok := value.(*string)
				if !ok {
					return fmt.Errorf("invalid data type for manager id")
				}
				return validation.Validate(v, is.UUIDv4)
			}),
		).Else(validation.Nil.Error("a non-staff role must not have a manager"))),
		validation.Field(&v.CreatedByID, validation.Required, is.UUIDv4),
		validation.Field(&v.EmployeeBiodata),
		validation.Field(&v.EmployeesEmergencyContacts),
	); err != nil {
		return err
	}

	for _, emergencyContact := range v.EmployeesEmergencyContacts {
		if v.FullName == emergencyContact.FullName || v.EmployeeBiodata.PhoneNumber == emergencyContact.PhoneNumber {
			return fmt.Errorf("cannot assign self as emergency contact")
		}
	}

	return nil
}

func (v Employee) ValidateLeave() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.ContractType,
			validation.Required,
			validation.By(func(value interface{}) error {
				v, ok := value.(ContractType)
				if !ok {
					return fmt.Errorf("invalid type for employees contract type")
				}

				if v == INTERN {
					return fmt.Errorf("interns are not allowed to request for leave")
				}

				return nil
			}),
		),
	)
}

func (v EmployeeDataHistoryLog) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Employee, validation.Required),
		validation.Field(&v.UpdatedBy, validation.Required),
		validation.Field(&v.UpdatedByID, validation.Required),
		validation.Field(&v.Changes, validation.Required, validation.NotNil.Error("a change must be committed")),
	)
}

func (v Employee) ValidateUpdateWorkInfo() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Status, validation.Required, validation.In(AVAILABLE, UNAVAILABLE, ON_LEAVE, RESIGNED)),
		validation.Field(&v.ContractType, validation.Required, validation.In(FULL_TIME, CONTRACT, INTERN)),
		validation.Field(&v.ManagerID, validation.When(v.Role.Code == "staff",
			validation.Required,
			validation.By(func(value interface{}) error {
				v, ok := value.(*string)
				if !ok {
					return fmt.Errorf("invalid data type for manager id")
				}
				return validation.Validate(v, is.UUIDv4)
			}),
		).Else(validation.Nil.Error("a non-staff role must not have a manager"))),
		validation.Field(&v.EmployeeDataHistoryLogs),
	)
}
