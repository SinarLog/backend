package mapper

import (
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

/*
*************************************************
REQUEST TO ENTITIES
*************************************************
*/
func MapClockInRequestToVO(req dto.ClockInRequest) vo.ClockInRequest {
	return vo.ClockInRequest{
		Credential: vo.Credential{
			OTP: req.OTP,
		},
		Loc: entity.Point{
			X: req.Long,
			Y: req.Lat,
		},
	}
}

func MapClockOutRequestToVO(req dto.ClockOutRequest) vo.ClockOutPayload {
	return vo.ClockOutPayload{
		Confirmation: req.Confirmation,
		Reason:       req.Reason,
		Loc: entity.Point{
			X: req.Long,
			Y: req.Lat,
		},
	}
}

/*
*************************************************
ENTITIES TO RESPONSE
*************************************************
*/
func MapAttendanceEntityToResponse(att entity.Attendance) dto.AttendanceResponse {
	res := dto.AttendanceResponse{
		EmployeeId:    att.EmployeeID,
		DoneForTheDay: att.DoneForTheDay,
		LateClockIn:   att.LateClockIn,
		EarlyClockOut: att.EarlyClockOut,
	}

	if !att.ClockInAt.IsZero() {
		res.ClockInAt = att.ClockInAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
	}

	if !att.ClockOutAt.IsZero() {
		res.ClockOutAt = att.ClockOutAt.In(utils.CURRENT_LOC).Format(time.RFC1123)
	}

	if att.ClockInLoc.X != float64(0) || att.ClockInLoc.Y != float64(0) {
		res.ClockInLoc = dto.LatLong{
			Long: att.ClockInLoc.X,
			Lat:  att.ClockInLoc.Y,
		}
	}

	if att.ClockOutLoc.X != float64(0) || att.ClockOutLoc.Y != float64(0) {
		res.ClockOutLoc = dto.LatLong{
			Long: att.ClockOutLoc.X,
			Lat:  att.ClockOutLoc.Y,
		}
	}

	return res
}

func MapMyAttendanceLogToResponse(att []entity.Attendance) []dto.MyAttendanceHistory {
	var res []dto.MyAttendanceHistory

	for _, v := range att {
		a := dto.MyAttendanceHistory{
			Date:          v.ClockInAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			ClockInAt:     v.ClockInAt.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
			ClockOutAt:    v.ClockOutAt.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
			DoneForTheDay: v.DoneForTheDay,
			ClockInLoc: dto.LatLong{
				Long: v.ClockInLoc.X,
				Lat:  v.ClockInLoc.Y,
			},
			ClockOutLoc: dto.LatLong{
				Long: v.ClockOutLoc.X,
				Lat:  v.ClockOutLoc.Y,
			},
			ClosedAutomatically: v.ClosedAutomatically != nil,
			LateClockIn:         v.LateClockIn,
			EarlyClockOut:       v.EarlyClockOut,
		}

		res = append(res, a)
	}

	return res
}

func MapEmployeesAttendanceLogToResponse(att []entity.Attendance) []dto.EmployeesAttendanceHistory {
	var res []dto.EmployeesAttendanceHistory

	for _, v := range att {
		a := dto.EmployeesAttendanceHistory{
			ID:            v.ID,
			Avatar:        v.Employee.Avatar,
			FullName:      v.Employee.FullName,
			Email:         v.Employee.Email,
			Position:      v.Employee.Job.Name,
			Date:          v.ClockInAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
			ClockInAt:     v.ClockInAt.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
			ClockOutAt:    v.ClockOutAt.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
			DoneForTheDay: v.DoneForTheDay,
			ClockInLoc: dto.LatLong{
				Long: v.ClockInLoc.X,
				Lat:  v.ClockInLoc.Y,
			},
			ClockOutLoc: dto.LatLong{
				Long: v.ClockOutLoc.X,
				Lat:  v.ClockOutLoc.Y,
			},
			ClosedAutomatically: v.ClosedAutomatically != nil,
			LateClockIn:         v.LateClockIn,
			EarlyClockOut:       v.EarlyClockOut,
		}

		res = append(res, a)
	}

	return res
}
