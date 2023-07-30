package vo

type CommonQuery struct {
	Pagination PaginationDTORequest
	TimeQuery  TimeQueryDTORequest
}

type AllEmployeeQuery struct {
	CommonQuery
	FullName string
	JobId    string
}

type IncomingLeaveProposals struct {
	CommonQuery
	Name string
}

type IncomingOvertimeSubmissionsQuery IncomingLeaveProposals

type MyOvertimeSubmissionsQuery struct {
	CommonQuery
	Status string
}

type LeaveProposalHistoryQuery struct {
	CommonQuery
	Status string
	Name   string
}

type HistoryAttendancesQuery struct {
	CommonQuery
	Early  bool
	Late   bool
	Closed bool
	Name   string
}
