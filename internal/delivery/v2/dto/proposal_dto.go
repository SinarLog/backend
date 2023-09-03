package dto

type IncomingLeaveProposalsForManagerResponse struct {
	ID          string `json:"id,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	RequestDate string `json:"requestDate,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Type        string `json:"type,omitempty"`
	Status      string `json:"status,omitempty"`
	Overflows   int    `json:"overflows,omitempty"`
}

type IncomingLeaveProposalsForHrResponse IncomingLeaveProposalsForManagerResponse

type IncomingLeaveProposalDetailForManagerResponse struct {
	ID          string                                      `json:"id,omitempty"`
	Avatar      string                                      `json:"avatar,omitempty"`
	FullName    string                                      `json:"fullName,omitempty"`
	Email       string                                      `json:"email,omitempty"`
	RequestDate string                                      `json:"requestDate,omitempty"`
	From        string                                      `json:"from,omitempty"`
	To          string                                      `json:"to,omitempty"`
	Reason      string                                      `json:"reason,omitempty"`
	Duration    int                                         `json:"duration,omitempty"`
	Type        string                                      `json:"type,omitempty"`
	Status      string                                      `json:"status,omitempty"`
	Attachment  string                                      `json:"attachment,omitempty"`
	Childs      []IncomingLeaveProposalChildsDetailResponse `json:"childs,omitempty"`
}

type IncomingLeaveProposalDetailForHrResponse struct {
	ID                string                                      `json:"id,omitempty"`
	Avatar            string                                      `json:"avatar,omitempty"`
	FullName          string                                      `json:"fullName,omitempty"`
	Email             string                                      `json:"email,omitempty"`
	IsManager         bool                                        `json:"isManager"`
	RequestDate       string                                      `json:"requestDate,omitempty"`
	From              string                                      `json:"from,omitempty"`
	To                string                                      `json:"to,omitempty"`
	Reason            string                                      `json:"reason,omitempty"`
	Duration          int                                         `json:"duration,omitempty"`
	Type              string                                      `json:"type,omitempty"`
	Status            string                                      `json:"status,omitempty"`
	Attachment        string                                      `json:"attachment,omitempty"`
	ApprovedByManager *bool                                       `json:"approvedByManager,omitempty"`
	ActionByManagerAt *string                                     `json:"actionByManagerAt,omitempty"`
	RejectionReason   string                                      `json:"rejectionReason,omitempty"`
	Manager           *BriefEmployeeListResponse                  `json:"manager,omitempty"`
	Childs            []IncomingLeaveProposalChildsDetailResponse `json:"childs,omitempty"`
}

type IncomingLeaveProposalChildsDetailResponse struct {
	ID                string  `json:"id,omitempty"`
	From              string  `json:"from,omitempty"`
	To                string  `json:"to,omitempty"`
	Reason            string  `json:"reason,omitempty"`
	Duration          int     `json:"duration,omitempty"`
	Type              string  `json:"type,omitempty"`
	Status            string  `json:"status,omitempty"`
	ApprovedByManager *bool   `json:"approvedByManager,omitempty"`
	ActionByManagerAt *string `json:"actionByManagerAt,omitempty"`
	RejectionReason   string  `json:"rejectionReason,omitempty"`
}
