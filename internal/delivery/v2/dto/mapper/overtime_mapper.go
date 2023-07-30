package mapper

import (
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

func overtimeStatusMapper(v entity.Overtime) string {
	if v.ApprovedByManager == nil {
		if v.ClosedAutomatically == nil {
			return "PENDING"
		} else {
			return "CLOSED"
		}
	} else if *v.ApprovedByManager {
		return "APPROVED"
	} else {
		return "REJECTED"
	}
}

/*
*************************************************
ENTITIES TO RESPONSE
*************************************************
*/
func MapOvertimeOnAttendanceReportToResponse(o entity.OvertimeOnAttendanceReport) dto.OvertimeOnAttendanceReportResponse {
	return dto.OvertimeOnAttendanceReportResponse{
		IsOvertime:                o.IsOvertime,
		IsOnHoliday:               o.IsOnHoliday,
		IsOvertimeLeakage:         o.IsOvertimeLeakage,
		IsOvertimeAvailable:       o.IsOvertimeAvailable,
		OvertimeDuration:          utils.SanitizeDuration(o.OvertimeDuration),
		OvertimeWeekTotalDuration: utils.SanitizeDuration(o.OvertimeWeekTotalDuration),
		OvertimeAcceptedDuration:  utils.SanitizeDuration(o.OvertimeAcceptedDuration),
		MaxAllowedDailyDuration:   utils.SanitizeDuration(o.MaxAllowedDailyDuration),
		MaxAllowedWeeklyDuration:  utils.SanitizeDuration(o.MaxAllowedWeeklyDuration),
	}
}

func MapIncomingOvertimeSubmissionsToResponse(ovs []entity.Overtime) []dto.IncomingOvertimeSubmissionsForManagerResponse {
	var res []dto.IncomingOvertimeSubmissionsForManagerResponse
	for _, v := range ovs {
		r := dto.IncomingOvertimeSubmissionsForManagerResponse{
			Id:       v.Id,
			Avatar:   v.Attendance.Employee.Avatar,
			FullName: v.Attendance.Employee.FullName,
			Date:     v.Attendance.ClockInAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Duration: utils.SanitizeDuration(time.Duration(v.Duration)),
			Status:   overtimeStatusMapper(v),
		}
		res = append(res, r)
	}

	return res
}

func MapOvertimeDetailToResponse(ov entity.Overtime) dto.OvertimeSubmissionDetailResponse {
	res := dto.OvertimeSubmissionDetailResponse{
		IncomingOvertimeSubmissionsForManagerResponse: dto.IncomingOvertimeSubmissionsForManagerResponse{
			Id:       ov.Id,
			Avatar:   ov.Attendance.Employee.Avatar,
			FullName: ov.Attendance.Employee.FullName,
			Date:     ov.Attendance.ClockInAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Duration: utils.SanitizeDuration(time.Duration(ov.Duration)),
		},
		Email:             ov.Attendance.Employee.Email,
		Reason:            ov.Reason,
		ApprovedByManager: ov.ApprovedByManager,
		RejectionReason:   ov.RejectionReason,
	}

	if ov.ApprovedByManager == nil {
		if ov.ClosedAutomatically == nil {
			res.Status = "PENDING"
		} else {
			res.Status = "CLOSED"
		}
	} else if *ov.ApprovedByManager {
		res.Status = "APPROVED"
	} else {
		res.Status = "REJECTED"
	}

	if ov.ClosedAutomatically != nil {
		res.ClosedAutomatically = *ov.ClosedAutomatically
	}

	if ov.ActionByManagerAt != nil {
		t := ov.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		res.ActionByManagerAt = &t
	}

	if ov.Manager != nil {
		res.Manager = &dto.BriefEmployeeListResponse{
			Id:       *ov.ManagerID,
			FullName: ov.Manager.FullName,
			Email:    ov.Manager.Email,
			Avatar:   ov.Manager.Avatar,
		}
	}

	return res
}

func MapMyOvertimeSubmissonToResponse(ovs []entity.Overtime) []dto.MyOvertimeSubmissionResponse {
	var res []dto.MyOvertimeSubmissionResponse

	for _, ov := range ovs {
		r := dto.MyOvertimeSubmissionResponse{
			Id:          ov.Id,
			RequestDate: ov.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Duration:    utils.SanitizeDuration(time.Duration(ov.Duration)),
		}

		if ov.ClosedAutomatically != nil {
			if *ov.ClosedAutomatically {
				r.Status = "CLOSED"
			}
		} else if ov.ApprovedByManager != nil {
			if *ov.ApprovedByManager {
				r.Status = "APPROVED"
			} else {
				r.Status = "REJECTED"
			}
		} else {
			r.Status = "PENDING"
		}

		res = append(res, r)
	}

	return res
}
