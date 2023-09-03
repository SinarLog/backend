package entity

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Overtime struct {
	BaseModelID

	AttendanceID  string `gorm:"type:uuid"`
	Duration      int
	Reason        string `gorm:"type:text"`
	AttachmentUrl string `gorm:"type:varchar(255)"`

	ManagerID           *string
	Manager             *Employee
	ApprovedByManager   *bool
	ActionByManagerAt   *time.Time
	RejectionReason     string
	ClosedAutomatically *bool

	Attendance Attendance

	BaseModelStamps
	BaseModelSoftDelete
}

func (v Overtime) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Duration, validation.Required, validation.Min(1)),
		validation.Field(&v.Reason, validation.Required, validation.Length(10, 1000).Error("overtime reason must be either 10 to 1000 characters long")),
		validation.Field(&v.ManagerID, validation.Required, validation.By(func(value interface{}) error {
			v, ok := value.(*string)
			if !ok {
				return fmt.Errorf("invalid data type for manager id")
			}

			if v == nil {
				return fmt.Errorf("only staff can have overtime submission. Manager id cannot be nil")
			}

			id := *v
			return validation.Validate(&id, validation.Required, is.UUIDv4.Error("manager id must be UUIDV4"))
		})),
	)
}

type OvertimeOnAttendanceReport struct {
	// Whether an attendance is an overtime
	IsOvertime bool `json:"isOvertime"`
	// Whether the attendance made is on holiday
	IsOnHoliday bool `json:"isOnHoliday"`
	// Whether the attendance duration is more than the allowed daily/weekly overtime duration
	IsOvertimeLeakage bool `json:"isOvertimeLeakage"`
	// Whether there could be made an overtime for that week
	IsOvertimeAvailable bool `json:"isOvertimeAvailable"`
	// Attendance's overtime duration
	OvertimeDuration time.Duration `json:"overtimeDuration"`
	// Overtime total duration for this week
	OvertimeWeekTotalDuration time.Duration `json:"overtimeWeeklyTotalDuration"`
	// Overtime accepted duration
	OvertimeAcceptedDuration time.Duration `json:"overtimeAcceptedDuration"`
	// Max allowed overtime daily duration
	MaxAllowedDailyDuration time.Duration `json:"maxAllowedDailyDuration,omitempty"`
	// Max allowed overtime weekly duration
	MaxAllowedWeeklyDuration time.Duration `json:"maxAllowedWeeklyDuration,omitempty"`
}

func (v OvertimeOnAttendanceReport) ShouldCreateOvertimeRecord() bool {
	if v.IsOvertime {
		if v.IsOvertimeAvailable {
			return true
		}
	}

	if v.IsOnHoliday {
		return true
	}

	return false
}
