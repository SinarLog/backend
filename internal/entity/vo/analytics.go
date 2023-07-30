package vo

type BriefLeaveAndAttendanceAnalytics struct {
	YearlyCount    int `json:"yearlyCount"`
	LateClockIns   int `json:"lateClockIns"`
	EarlyClockOuts int `json:"earlyClockOuts"`
	UnpaidCount    int `json:"unpaidCount"`
}

type HrDashboardAnalytics struct {
	TotalEmployees               int64  `json:"totalEmployees"`
	LateClockIns                 int64  `json:"lateClockIns"`
	EarlyClockOuts               int64  `json:"earlyClockOuts"`
	ApprovedUnpaidLeaves         int64  `json:"approvedUnpaidLeaves"`
	ApprovedAnnualMarriageLeaves int64  `json:"approvedAnnualMarriageLeaves"`
	SickLeaves                   int64  `json:"sickLeaves"`
	Month                        string `json:"currentMonth"`
}
