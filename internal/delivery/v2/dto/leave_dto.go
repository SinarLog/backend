package dto

type MyLeaveRequestListsResponse struct {
	ID          string `json:"id,omitempty"`
	RequestDate string `json:"requestDate,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Status      string `json:"status,omitempty"`
	LeaveType   string `json:"leaveType,omitempty"`
}

type LeaveRequest struct {
	ID          string `json:"id,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Type        string `json:"type,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	RequestDate string `json:"requestDate,omitempty"`
	Status      string `json:"status,omitempty"`
}

type MyLeaveRequestDetailResponse struct {
	ID            string                     `json:"id,omitempty"`
	RequestDate   string                     `json:"requestDate,omitempty"`
	From          string                     `json:"from,omitempty"`
	To            string                     `json:"to,omitempty"`
	Duration      int                        `json:"duration,omitempty"`
	Type          string                     `json:"type,omitempty"`
	Reason        string                     `json:"reason,omitempty"`
	Status        string                     `json:"status"`
	AttachmentUrl string                     `json:"attachmentUrl,omitempty"`
	Manager       *BriefEmployeeListResponse `json:"manager,omitempty"`
	Hr            *BriefEmployeeListResponse `json:"hr,omitempty"`

	ApprovedByHr      *bool   `json:"approvedByHr,omitempty"`
	ApprovedByManager *bool   `json:"approvedByManager,omitempty"`
	ActionByHrAt      *string `json:"actionByHrAt,omitempty"`
	ActionByManagerAt *string `json:"actionByManagerAt,omitempty"`
	RejectionReason   string  `json:"rejectionReason,omitempty"`

	Parent              *LeaveRequest  `json:"parent,omitempty"`
	Childs              []LeaveRequest `json:"childs,omitempty"`
	ClosedAutomatically *bool          `json:"closedAutomatically,omitempty"`
}

type LeaveRequestDetailResponse struct {
	ID            string                     `json:"id,omitempty"`
	Avatar        string                     `json:"avatar,omitempty"`
	FullName      string                     `json:"fullName,omitempty"`
	Email         string                     `json:"email,omitempty"`
	Duration      int                        `json:"duration,omitempty"`
	RequestDate   string                     `json:"requestDate,omitempty"`
	From          string                     `json:"from,omitempty"`
	To            string                     `json:"to,omitempty"`
	Type          string                     `json:"type,omitempty"`
	Reason        string                     `json:"reason,omitempty"`
	Status        string                     `json:"status"`
	AttachmentUrl string                     `json:"attachmentUrl,omitempty"`
	Manager       *BriefEmployeeListResponse `json:"manager,omitempty"`
	Hr            *BriefEmployeeListResponse `json:"hr,omitempty"`

	ApprovedByHr      *bool   `json:"approvedByHr,omitempty"`
	ApprovedByManager *bool   `json:"approvedByManager,omitempty"`
	ActionByHrAt      *string `json:"actionByHrAt,omitempty"`
	ActionByManagerAt *string `json:"actionByManagerAt,omitempty"`
	RejectionReason   string  `json:"rejectionReason,omitempty"`

	Parent              *LeaveRequest                `json:"parent,omitempty"`
	Childs              []LeaveRequestDetailResponse `json:"childs,omitempty"`
	ClosedAutomatically *bool                        `json:"closedAutomatically,omitempty"`
}

type LeaveRequestReportExcessResponse struct {
	Type  string `json:"type,omitempty"`
	Quota int    `json:"quota,omitempty"`
}
type LeaveRequestReportResponse struct {
	IsLeaveLeakage                 bool                               `json:"isLeaveLeakage"`
	ExcessLeaveDuration            int                                `json:"excessLeaveDuration"`
	RequestType                    string                             `json:"requestType,omitempty"`
	RemainingQuotaForRequestedType int                                `json:"remainingQuotaForRequestedType"`
	Availables                     []LeaveRequestReportExcessResponse `json:"availables,omitempty"`
}

type LeaveDecision struct {
	Parent    LeaveRequest             `json:"parent,omitempty"`
	Overflows []LeaveOverflowsDecision `json:"overflows,omitempty"`
}

type LeaveOverflowsDecision struct {
	Type  string `json:"type,omitempty"`
	Count int    `json:"count,omitempty"`
}
