package entity

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"sinarlog.com/internal/utils"
)

type Attendance struct {
	BaseModelId

	EmployeeID          string `gorm:"type:uuid"`
	Employee            Employee
	ClockInAt           time.Time `gorm:"default:now()"`
	ClockOutAt          time.Time
	DoneForTheDay       bool
	ClockInLoc          Point
	ClockOutLoc         Point
	LateClockIn         bool
	EarlyClockOut       bool
	ClosedAutomatically *bool

	Overtime *Overtime

	BaseModelStamps
	BaseModelSoftDelete
}

// ValidateClockIn validates two things:
// 1. The clock in time must not be greater than the office
// end time minus the accepted attandance interval.
// 2. The location point must not be nil and is a valid (lat, long)
// data type.
func (v Attendance) ValidateClockIn(config Configuration) error {
	var errs error

	dur, err := time.ParseDuration(config.AcceptanceAttendanceInterval)
	if err != nil {
		errs = utils.AddError(errs, err)
	}

	officeEndTime := time.Date(
		v.ClockInAt.Year(),
		v.ClockInAt.Month(),
		v.ClockInAt.Day(),
		config.OfficeEndTime.Hour(),
		config.OfficeEndTime.Minute(),
		config.OfficeEndTime.Second(),
		config.OfficeEndTime.Nanosecond(),
		utils.CURRENT_LOC,
	)

	if v.ClockInAt.After(officeEndTime.Add(-dur)) {
		errs = utils.AddError(errs, fmt.Errorf("clock in after office's end time is not allowed"))
	}

	if err := v.ClockInLoc.Validate(); err != nil {
		errs = utils.AddError(errs, err)
	}

	x := fmt.Sprintf("%f", v.ClockInLoc.X)
	y := fmt.Sprintf("%f", v.ClockInLoc.Y)

	if err := validation.Validate(&x, validation.Required, is.Float, is.Longitude); err != nil {
		errs = utils.AddError(errs, err)
	}

	if err := validation.Validate(&y, validation.Required, is.Float, is.Latitude); err != nil {
		errs = utils.AddError(errs, err)
	}

	return errs
}

// IsLateClockIn returns whether the attendance is a late clock in
// according to the passed office configuration entity or the day
// the clock in is made.
func (v Attendance) IsLateClockIn(config Configuration) bool {
	if v.ClockInAt.Weekday() != time.Saturday && v.ClockInAt.Weekday() != time.Sunday {
		interval, _ := time.ParseDuration(config.AcceptanceAttendanceInterval)

		lateTime := time.Date(
			v.ClockInAt.Year(),
			v.ClockInAt.Month(),
			v.ClockInAt.Day(),
			config.OfficeStartTime.Hour(),
			config.OfficeStartTime.Minute(),
			config.OfficeStartTime.Second(),
			config.OfficeStartTime.Nanosecond(),
			utils.CURRENT_LOC,
		).Add(interval)

		return v.ClockInAt.After(lateTime)
	}

	return false
}

func (v Attendance) ValidateClockOut() error {
	var errs error

	if err := v.ClockOutLoc.Validate(); err != nil {
		errs = utils.AddError(errs, err)
	}

	x := fmt.Sprintf("%f", v.ClockOutLoc.X)
	y := fmt.Sprintf("%f", v.ClockOutLoc.Y)

	if err := validation.Validate(&x, validation.Required, is.Float, is.Longitude); err != nil {
		errs = utils.AddError(errs, err)
	}

	if err := validation.Validate(&y, validation.Required, is.Float, is.Latitude); err != nil {
		errs = utils.AddError(errs, err)
	}

	if v.ClockInLoc.DistanceTo(v.ClockOutLoc) > 5000 {
		errs = utils.AddError(errs, fmt.Errorf("the distance between clock in and clock out location is too large. Please be around 5km around your clock in location"))
	}

	return errs
}

func (v Attendance) IsEarlyClockOut(config Configuration) bool {
	if v.ClockInAt.Weekday() != time.Saturday && v.ClockInAt.Weekday() != time.Sunday {
		interval, _ := time.ParseDuration(config.AcceptanceAttendanceInterval)

		lateTime := time.Date(
			v.ClockOutAt.Year(),
			v.ClockOutAt.Month(),
			v.ClockOutAt.Day(),
			config.OfficeEndTime.Hour(),
			config.OfficeEndTime.Minute(),
			config.OfficeEndTime.Second(),
			config.OfficeEndTime.Nanosecond(),
			utils.CURRENT_LOC,
		).Add(-interval)

		return v.ClockOutAt.Before(lateTime)
	}

	return false
}
