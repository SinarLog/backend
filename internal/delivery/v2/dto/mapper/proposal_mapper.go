package mapper

import (
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

/*
*************************************************
ENTITIES TO RESPONSE
*************************************************
*/
func MapIncomingLeaveProposalsForManagerResponse(leaves []entity.Leave) []dto.IncomingLeaveProposalsForManagerResponse {
	var res []dto.IncomingLeaveProposalsForManagerResponse

	for _, v := range leaves {
		d := dto.IncomingLeaveProposalsForManagerResponse{
			ID:          v.ID,
			Avatar:      v.Employee.Avatar,
			FullName:    v.Employee.FullName,
			RequestDate: v.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			From:        v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:          v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Duration:    utils.CountNumberOfWorkingDays(v.From, v.To),
			Type:        v.Type.String(),
			Status:      LeaveStatusMapper(v),
		}

		if v.Childs != nil {
			d.Overflows = len(v.Childs)
		}

		res = append(res, d)
	}

	return res
}

func MapIncomingLeaveProposalsForHrResponse(leaves []entity.Leave) []dto.IncomingLeaveProposalsForHrResponse {
	var res []dto.IncomingLeaveProposalsForHrResponse

	for _, v := range leaves {
		d := dto.IncomingLeaveProposalsForHrResponse{
			ID:          v.ID,
			Avatar:      v.Employee.Avatar,
			FullName:    v.Employee.FullName,
			RequestDate: v.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			From:        v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:          v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Duration:    utils.CountNumberOfWorkingDays(v.From, v.To),
			Type:        v.Type.String(),
			Status:      LeaveStatusMapper(v),
		}

		if v.Childs != nil {
			d.Overflows = len(v.Childs)
		}

		res = append(res, d)
	}

	return res
}

func MapIncomingLeaveProposalDetailForManagerResponse(leave entity.Leave) dto.IncomingLeaveProposalDetailForManagerResponse {
	res := dto.IncomingLeaveProposalDetailForManagerResponse{
		ID:          leave.ID,
		Avatar:      leave.Employee.Avatar,
		FullName:    leave.Employee.FullName,
		Email:       leave.Employee.Email,
		RequestDate: leave.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		From:        leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
		To:          leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
		Reason:      leave.Reason,
		Duration:    utils.CountNumberOfWorkingDays(leave.From, leave.To),
		Type:        leave.Type.String(),
		Status:      "PENDING",
		Attachment:  leave.AttachmentUrl,
	}

	for _, v := range leave.Childs {
		d := dto.IncomingLeaveProposalChildsDetailResponse{
			ID:       v.ID,
			From:     v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:       v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Reason:   v.Reason,
			Duration: utils.CountNumberOfWorkingDays(v.From, v.To),
			Type:     v.Type.String(),
			Status:   "PENDING",
		}

		res.Childs = append(res.Childs, d)
	}

	return res
}

func MapIncomingLeaveProposalDetailForHrResponse(leave entity.Leave) dto.IncomingLeaveProposalDetailForHrResponse {
	res := dto.IncomingLeaveProposalDetailForHrResponse{
		ID:                leave.ID,
		Avatar:            leave.Employee.Avatar,
		FullName:          leave.Employee.FullName,
		Email:             leave.Employee.Email,
		IsManager:         leave.Employee.ManagerID == nil,
		RequestDate:       leave.CreatedAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		From:              leave.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
		To:                leave.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
		Reason:            leave.Reason,
		Duration:          utils.CountNumberOfWorkingDays(leave.From, leave.To),
		Type:              leave.Type.String(),
		Status:            LeaveStatusMapper(leave),
		Attachment:        leave.AttachmentUrl,
		ApprovedByManager: leave.ApprovedByManager,
		RejectionReason:   leave.RejectionReason,
	}

	actionByManagerAt := leave.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
	res.ActionByManagerAt = &actionByManagerAt

	if leave.Manager != nil {
		res.Manager = &dto.BriefEmployeeListResponse{
			ID:       leave.Manager.ID,
			FullName: leave.Manager.FullName,
			Email:    leave.Manager.Email,
			Avatar:   leave.Manager.Avatar,
		}
	}

	for _, v := range leave.Childs {
		c := dto.IncomingLeaveProposalChildsDetailResponse{
			ID:                v.ID,
			From:              v.From.In(utils.CURRENT_LOC).Format(time.DateOnly),
			To:                v.To.In(utils.CURRENT_LOC).Format(time.DateOnly),
			Reason:            v.Reason,
			Duration:          utils.CountNumberOfWorkingDays(v.From, v.To),
			Type:              v.Type.String(),
			Status:            LeaveStatusMapper(v),
			ApprovedByManager: v.ApprovedByManager,
			RejectionReason:   v.RejectionReason,
		}
		actionByManagerAt := v.ActionByManagerAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
		c.ActionByManagerAt = &actionByManagerAt

		res.Childs = append(res.Childs, c)
	}

	return res
}
