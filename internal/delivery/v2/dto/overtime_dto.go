package dto

type OvertimeOnAttendanceReportResponse struct {
	// Whether an attendance is an overtime
	IsOvertime bool `json:"isOvertime"`
	// Whether the attendance made is on holiday
	IsOnHoliday bool `json:"isOnHoliday"`
	// Whether the attendance duration is more than the allowed daily/weekly overtime duration
	IsOvertimeLeakage bool `json:"isOvertimeLeakage"`
	// Whether there could be made an overtime for that week
	IsOvertimeAvailable bool `json:"isOvertimeAvailable"`
	// Attendance's overtime duration
	OvertimeDuration string `json:"overtimeDuration"`
	// Overtime total duration for this week
	OvertimeWeekTotalDuration string `json:"overtimeWeeklyTotalDuration"`
	// Overtime accepted duration
	OvertimeAcceptedDuration string `json:"overtimeAcceptedDuration"`
	// Max allowed overtime daily duration
	MaxAllowedDailyDuration string `json:"maxAllowedDailyDuration,omitempty"`
	// Max allowed overtime weekly duration
	MaxAllowedWeeklyDuration string `json:"maxAllowedWeeklyDuration,omitempty"`
}

type IncomingOvertimeSubmissionsForManagerResponse struct {
	ID       string `json:"id,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	FullName string `json:"fullName,omitempty"`
	Date     string `json:"date,omitempty"`
	Duration string `json:"duration,omitempty"`
	Status   string `json:"status,omitempty"`
}

type OvertimeSubmissionDetailResponse struct {
	IncomingOvertimeSubmissionsForManagerResponse
	Email               string                     `json:"email,omitempty"`
	Reason              string                     `json:"reason,omitempty"`
	ApprovedByManager   *bool                      `json:"approvedByManager,omitempty"`
	ActionByManagerAt   *string                    `json:"actionByManagerAt,omitempty"`
	RejectionReason     string                     `json:"rejectionReason,omitempty"`
	Manager             *BriefEmployeeListResponse `json:"manager,omitempty"`
	ClosedAutomatically bool                       `json:"closedAutomatically,omitempty"`
}

type MyOvertimeSubmissionResponse struct {
	ID          string `json:"id,omitempty"`
	RequestDate string `json:"requestDate,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Status      string `json:"status,omitempty"`
}
