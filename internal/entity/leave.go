package entity

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"sinarlog.com/internal/utils"
)

type LeaveType string

const (
	ANNUAL   LeaveType = "ANNUAL"
	UNPAID   LeaveType = "UNPAID"
	SICK     LeaveType = "SICK"
	MARRIAGE LeaveType = "MARRIAGE"
)

func (l LeaveType) String() string {
	switch l {
	case ANNUAL:
		return "ANNUAL"
	case UNPAID:
		return "UNPAID"
	case SICK:
		return "SICK"
	case MARRIAGE:
		return "MARRIAGE"
	default:
		return ""
	}
}

type Leave struct {
	BaseModelID

	EmployeeID    string `gorm:"type:uuid"`
	Employee      Employee
	From          time.Time
	To            time.Time
	Type          LeaveType `gorm:"type:varchar(100)"`
	Reason        string    `gorm:"type:text"`
	AttachmentUrl string    `gorm:"type:varchar(255)"`

	// A parent leave contains the original leave request.
	Parent   *Leave
	ParentID *string `gorm:"type:uuid;default:null"`
	// A childs leave contains the overflowed excess of the
	// original leave request. For example, requesting a leave
	// of type ANNUAL with 14 days duration meanwhile I have only
	// 12 ANNUAL quota left. Hence, I can overflow it to an
	// UNPAID leave request of 2 days, with my parent leave request
	// of ANNUAL of 12 days.
	Childs []Leave `gorm:"foreignKey:ParentID"`

	ManagerID           *string `gorm:"type:uuid;default:null"`
	Manager             *Employee
	HrID                *string `gorm:"type:uuid;default:null"`
	Hr                  *Employee
	ApprovedByManager   *bool
	ApprovedByHr        *bool
	ActionByManagerAt   *time.Time
	ActionByHrAt        *time.Time
	RejectionReason     string `gorm:"type:text"`
	ClosedAutomatically *bool

	BaseModelStamps
	BaseModelSoftDelete
}

func (v Leave) Validate() error {
	err := validation.ValidateStruct(&v,
		validation.Field(&v.From,
			validation.Required.Error("leave rquest start date is required"),
			validation.Max(v.To).Error("leave request starting date must not be before than ending date.")),
		validation.Field(&v.To,
			validation.Required.Error("leave request end date is required"),
			validation.Min(v.From).Error("leave request ending date must not be after starting date.")),
		validation.Field(&v.Reason,
			validation.Required.Error("leave request reason is required"),
			validation.Length(20, 1000).Error("leave request length must be between 20 and 1000")),
		validation.Field(&v.Type, validation.Required.Error("leave request type is required")),
	)
	if err != nil {
		return err
	}

	// Checks whether the selected days all holidays
	numDays := utils.CountNumberOfDays(v.From, v.To)
	haveWeekDay := false
	for i := 0; i < numDays; i++ {
		day := v.From.Add(time.Duration(i) * 24 * time.Hour)
		if day.Weekday() != time.Sunday && day.Weekday() != time.Saturday {
			haveWeekDay = true
			break
		}
	}

	if haveWeekDay {
		return nil
	}

	return fmt.Errorf("all selected days are holidays")
}

type LeaveReport struct {
	// Whether the leaves quota exceeds the quota according to the type
	IsLeaveLeakage bool
	// The excessive leave duration as days
	ExcessLeaveDuration int

	// Stores the initial request type
	RequestType LeaveType
	// Stores the remaining quota for the request type
	RemainingQuotaForRequestedType int

	// The available excess types to overflow the leakage
	AvailableExcessTypes []LeaveType
	// The available excess quota to overflow the leakage
	// NOTES: Make unpaid count to 10 max
	AvailableExcessQuotas []int
}
