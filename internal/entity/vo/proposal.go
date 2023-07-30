package vo

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"sinarlog.com/internal/entity"
)

type UserLeaveDecision struct {
	Parent    entity.Leave
	Overflows []LeaveOverflowsDecision
}

type LeaveOverflowsDecision struct {
	Type  entity.LeaveType
	Count int
}

func (v UserLeaveDecision) ValidateExcessSumOfDays(report entity.LeaveReport) error {
	var sum int
	for _, v := range v.Overflows {
		sum += v.Count
	}

	if sum != report.ExcessLeaveDuration {
		return fmt.Errorf("the excessed leave durations is not equal to the required excess duration")
	}

	return nil
}

type LeaveAction struct {
	Id       string        `json:"id,omitempty" binding:"required"`
	Approved bool          `json:"approved,omitempty"`
	Reason   string        `json:"reason,omitempty"`
	Childs   []LeaveAction `json:"childs,omitempty"`
}

func (v LeaveAction) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Id,
			validation.Required.Error("id field is required"),
			is.UUIDv4.Error("id must be a uuid"),
		),
		validation.Field(&v.Childs),
	)
}

type OvertimeSubmissionAction struct {
	Id       string `json:"id,omitempty" binding:"required"`
	Approved bool   `json:"approved,omitempty"`
	Reason   string `json:"reason,omitempty"`
}
