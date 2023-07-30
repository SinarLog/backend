package dto

type ClockInRequest struct {
	OTP  string  `json:"otp,omitempty"`
	Lat  float64 `json:"lat,omitempty"`
	Long float64 `json:"long,omitempty"`
}

type ClockOutRequest struct {
	Confirmation bool    `json:"confirmation"`
	Lat          float64 `json:"lat,omitempty" binding:"required"`
	Long         float64 `json:"long,omitempty" binding:"required"`
	Reason       string  `json:"reason,omitempty"`
}

type AttendanceResponse struct {
	EmployeeId    string  `json:"employeeId,omitempty"`
	ClockInAt     string  `json:"clockInAt,omitempty"`
	ClockOutAt    string  `json:"clockOutAt,omitempty"`
	DoneForTheDay bool    `json:"doneForTheDay"`
	ClockInLoc    LatLong `json:"clockInLoc,omitempty"`
	ClockOutLoc   LatLong `json:"clockOutLoc,omitempty"`
	LateClockIn   bool    `json:"lateClockIn,omitempty"`
	EarlyClockOut bool    `json:"earlyClockOut,omitempty"`
}

type LatLong struct {
	Long float64 `json:"long,omitempty"`
	Lat  float64 `json:"lat,omitempty"`
}

type MyAttendanceHistory struct {
	Date                string  `json:"date,omitempty"`
	ClockInAt           string  `json:"clockInAt,omitempty"`
	ClockOutAt          string  `json:"clockOutAt,omitempty"`
	DoneForTheDay       bool    `json:"doneForTheDay,omitempty"`
	ClockInLoc          LatLong `json:"clockInLoc,omitempty"`
	ClockOutLoc         LatLong `json:"clockOutLoc,omitempty"`
	ClosedAutomatically bool    `json:"closedAutomatically,omitempty"`
	LateClockIn         bool    `json:"lateClockIn"`
	EarlyClockOut       bool    `json:"earlyClockOut"`
}

type EmployeesAttendanceHistory struct {
	Id                  string  `json:"id,omitempty"`
	Avatar              string  `json:"avatar,omitempty"`
	FullName            string  `json:"fullName,omitempty"`
	Email               string  `json:"email,omitempty"`
	Position            string  `json:"position,omitempty"`
	Date                string  `json:"date,omitempty"`
	ClockInAt           string  `json:"clockInAt,omitempty"`
	ClockOutAt          string  `json:"clockOutAt,omitempty"`
	DoneForTheDay       bool    `json:"doneForTheDay"`
	ClockInLoc          LatLong `json:"clockInLoc,omitempty"`
	ClockOutLoc         LatLong `json:"clockOutLoc,omitempty"`
	ClosedAutomatically bool    `json:"closedAutomatically,omitempty"`
	LateClockIn         bool    `json:"lateClockIn"`
	EarlyClockOut       bool    `json:"earlyClockOut"`
}
