package mapper

import (
	"fmt"
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

func LeaveStatusMapper(v entity.Leave) string {
	var l string

	if v.ApprovedByHr != nil && v.ApprovedByManager != nil {
		if *v.ApprovedByManager && *v.ApprovedByHr {
			l = "APPROVED"
		} else {
			l = "REJECTED"
		}
	} else if v.ApprovedByManager != nil {
		if !*v.ApprovedByManager {
			l = "REJECTED"
		} else {
			l = "PENDING"
		}
	} else {
		l = "PENDING"
	}

	if v.ClosedAutomatically != nil {
		l = "CLOSED"
	}

	return l
}

/*
*************************************************
REQUEST TO ENTITIES
*************************************************
*/
func MapLeaveRequestToDomain(req dto.LeaveRequest) (entity.Leave, error) {
	res := entity.Leave{
		Reason: req.Reason,
		Type:   entity.LeaveType(req.Type),
	}

	from, err := time.Parse(time.DateOnly, req.From)
	if err != nil {
		return res, fmt.Errorf("invalid start date format")
	}
	res.From = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, utils.CURRENT_LOC)
	fmt.Println(from)

	to, err := time.Parse(time.DateOnly, req.To)
	if err != nil {
		return res, fmt.Errorf("invalide end date format")
	}
	res.To = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, utils.CURRENT_LOC)

	return res, nil
}

func MapLeaveDecisionToVO(req dto.LeaveDecision) (vo.UserLeaveDecision, error) {
	var res vo.UserLeaveDecision

	parent, err := MapLeaveRequestToDomain(req.Parent)
	if err != nil {
		return res, err
	}
	res.Parent = parent

	for _, v := range req.Overflows {
		res.Overflows = append(res.Overflows, vo.LeaveOverflowsDecision{
			Type:  entity.LeaveType(v.Type),
			Count: v.Count,
		})
	}

	return res, nil
}

/*
*************************************************
ENTITIES TO RESPONSE
*************************************************
*/
func MapMyLeaveRequestListToResponse(leaves []entity.Leave) []dto.MyLeaveRequestListsResponse {
	var res []dto.MyLeaveRequestListsResponse

	for _, v := range leaves {
		l := dto.MyLeaveRequestListsResponse{
			ID:          v.ID,
			RequestDate: v.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			From:        v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:          v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			LeaveType:   v.Type.String(),
			Duration:    utils.CountNumberOfWorkingDays(v.From, v.To),
		}

		l.Status = LeaveStatusMapper(v)
		res = append(res, l)
	}

	return res
}

func MapMyLeaveRequestDetailToResponse(leave entity.Leave) dto.MyLeaveRequestDetailResponse {
	res := dto.MyLeaveRequestDetailResponse{
		ID:                  leave.ID,
		RequestDate:         leave.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		From:                leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
		To:                  leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
		Type:                leave.Type.String(),
		Reason:              leave.Reason,
		Duration:            utils.CountNumberOfWorkingDays(leave.From, leave.To),
		AttachmentUrl:       leave.AttachmentUrl,
		ApprovedByHr:        leave.ApprovedByHr,
		ApprovedByManager:   leave.ApprovedByManager,
		RejectionReason:     leave.RejectionReason,
		ClosedAutomatically: leave.ClosedAutomatically,
	}

	res.Status = LeaveStatusMapper(leave)

	if leave.ActionByHrAt != nil {
		s := leave.ActionByHrAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		res.ActionByHrAt = &s
	}

	if leave.ActionByManagerAt != nil {
		s := leave.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		res.ActionByManagerAt = &s
	}

	if leave.Parent != nil {
		res.Parent = &dto.LeaveRequest{
			ID:          *leave.ParentID,
			From:        leave.Parent.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:          leave.Parent.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Type:        leave.Parent.Type.String(),
			Reason:      leave.Parent.Reason,
			Duration:    utils.CountNumberOfWorkingDays(leave.Parent.From, leave.Parent.To),
			RequestDate: leave.Parent.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		}

		res.Parent.Status = LeaveStatusMapper(*leave.Parent)
	}

	if leave.Childs != nil {
		for _, v := range leave.Childs {
			c := dto.LeaveRequest{
				ID:       v.ID,
				From:     v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
				To:       v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
				Type:     v.Type.String(),
				Duration: utils.CountNumberOfWorkingDays(v.From, v.To),
				Reason:   v.Reason,
			}
			c.Status = LeaveStatusMapper(v)
			res.Childs = append(res.Childs, c)
		}
	}

	if leave.Manager != nil {
		res.Manager = &dto.BriefEmployeeListResponse{
			ID:       leave.Manager.ID,
			FullName: leave.Manager.FullName,
			Email:    leave.Manager.Email,
			Avatar:   leave.Manager.Avatar,
		}
	}

	if leave.Hr != nil {
		res.Hr = &dto.BriefEmployeeListResponse{
			ID:       leave.Hr.ID,
			FullName: leave.Hr.FullName,
			Email:    leave.Hr.Email,
			Avatar:   leave.Hr.Avatar,
		}
	}

	return res
}

func MapLeaveRequestReportToResponse(report entity.LeaveReport) dto.LeaveRequestReportResponse {
	res := dto.LeaveRequestReportResponse{
		IsLeaveLeakage:                 report.IsLeaveLeakage,
		ExcessLeaveDuration:            report.ExcessLeaveDuration,
		RequestType:                    report.RequestType.String(),
		RemainingQuotaForRequestedType: report.RemainingQuotaForRequestedType,
	}

	for i := 0; i < len(report.AvailableExcessTypes); i++ {
		res.Availables = append(res.Availables, dto.LeaveRequestReportExcessResponse{
			Type:  report.AvailableExcessTypes[i].String(),
			Quota: report.AvailableExcessQuotas[i],
		})
	}

	return res
}

func MapLeaveRequestDetailToResponse(leave entity.Leave) dto.LeaveRequestDetailResponse {
	res := dto.LeaveRequestDetailResponse{
		ID:                  leave.ID,
		RequestDate:         leave.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		From:                leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
		To:                  leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
		Type:                leave.Type.String(),
		Avatar:              leave.Employee.Avatar,
		FullName:            leave.Employee.FullName,
		Email:               leave.Employee.Email,
		Duration:            utils.CountNumberOfWorkingDays(leave.From, leave.To),
		Reason:              leave.Reason,
		AttachmentUrl:       leave.AttachmentUrl,
		ApprovedByHr:        leave.ApprovedByHr,
		ApprovedByManager:   leave.ApprovedByManager,
		RejectionReason:     leave.RejectionReason,
		ClosedAutomatically: leave.ClosedAutomatically,
	}

	res.Status = LeaveStatusMapper(leave)

	if leave.ActionByHrAt != nil {
		s := leave.ActionByHrAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		res.ActionByHrAt = &s
	}

	if leave.ActionByManagerAt != nil {
		s := leave.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		res.ActionByManagerAt = &s
	}

	if leave.Parent != nil {
		res.Parent = &dto.LeaveRequest{
			ID:          *leave.ParentID,
			From:        leave.Parent.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:          leave.Parent.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Type:        leave.Parent.Type.String(),
			Reason:      leave.Parent.Reason,
			RequestDate: leave.Parent.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		}

		res.Parent.Status = LeaveStatusMapper(*leave.Parent)
	}

	if leave.Childs != nil {
		for _, v := range leave.Childs {
			c := dto.LeaveRequestDetailResponse{
				ID:                  v.ID,
				RequestDate:         v.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
				From:                v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
				To:                  v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
				Duration:            utils.CountNumberOfWorkingDays(v.From, v.To),
				Type:                v.Type.String(),
				Reason:              v.Reason,
				Status:              LeaveStatusMapper(v),
				AttachmentUrl:       v.AttachmentUrl,
				ApprovedByHr:        v.ApprovedByHr,
				ApprovedByManager:   v.ApprovedByManager,
				RejectionReason:     v.RejectionReason,
				ClosedAutomatically: v.ClosedAutomatically,
			}

			if v.ActionByHrAt != nil {
				h := v.ActionByHrAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
				c.ActionByHrAt = &h
			}

			if v.ActionByManagerAt != nil {
				m := v.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
				c.ActionByManagerAt = &m
			}

			res.Childs = append(res.Childs, c)
		}
	}

	if leave.Manager != nil {
		res.Manager = &dto.BriefEmployeeListResponse{
			ID:       leave.Manager.ID,
			FullName: leave.Manager.FullName,
			Email:    leave.Manager.Email,
			Avatar:   leave.Manager.Avatar,
		}
	}

	if leave.Hr != nil {
		res.Hr = &dto.BriefEmployeeListResponse{
			ID:       leave.Hr.ID,
			FullName: leave.Hr.FullName,
			Email:    leave.Hr.Email,
			Avatar:   leave.Hr.Avatar,
		}
	}

	return res
}
