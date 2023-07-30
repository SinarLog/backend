package dto

type GlobalConfigResponse struct {
	StartTime                    string `json:"startTime,omitempty"`
	EndTime                      string `json:"endTime,omitempty"`
	AcceptanceAttendanceInterval string `json:"acceptedAttendanceInterval"`
	AcceptanceLeaveInterval      int    `json:"acceptedLeaveInterval"`
	DefaultYearlyQuota           int    `json:"defaultYearlyQuota"`
	DefaultMarriageQuota         int    `json:"defaultMarriageQuota"`
}

type UpdateConfigRequest struct {
	StartTimeHour                int    `json:"startTimeHour"`
	StartTimeMinute              int    `json:"startTimeMinute"`
	EndTimeHour                  int    `json:"endTimeHour"`
	EndTimeMinute                int    `json:"endTimeMinute"`
	AcceptanceAttendanceInterval string `json:"acceptanceAttendanceInterval"`
	AcceptanceLeaveInterval      int    `json:"acceptanceLeaveInterval"`
	DefaultYearlyQuota           int    `json:"defaultYearlyQuota"`
	DefaultMarriageQuota         int    `json:"defaultMarriageQuota"`
}

type ConfigChangesLogsResponse struct {
	Id          string                    `json:"id,omitempty"`
	UpdatedBy   BriefEmployeeListResponse `json:"updatedBy,omitempty"`
	Changes     map[string]any            `json:"changes,omitempty"`
	UpdatedAt   string                    `json:"updatedAt,omitempty"`
	WhenApplied string                    `json:"whenApplied,omitempty"`
}
