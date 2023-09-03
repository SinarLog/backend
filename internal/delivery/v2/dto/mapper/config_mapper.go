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
func MapConfigResponse(data entity.Configuration) dto.GlobalConfigResponse {
	return dto.GlobalConfigResponse{
		StartTime:                    data.OfficeStartTime.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
		EndTime:                      data.OfficeEndTime.In(utils.CURRENT_LOC).Format(time.TimeOnly)[:5],
		AcceptanceAttendanceInterval: data.AcceptanceAttendanceInterval,
		AcceptanceLeaveInterval:      data.AcceptanceLeaveInterval,
		DefaultYearlyQuota:           data.DefaultYearlyQuota,
		DefaultMarriageQuota:         data.DefaultMarriageQuota,
	}
}

func MapConfigDetailResponse(data entity.Configuration) dto.UpdateConfigRequest {
	return dto.UpdateConfigRequest{
		StartTimeHour:                data.OfficeStartTime.Hour(),
		StartTimeMinute:              data.OfficeStartTime.Minute(),
		EndTimeHour:                  data.OfficeEndTime.Hour(),
		EndTimeMinute:                data.OfficeEndTime.Minute(),
		AcceptanceAttendanceInterval: data.AcceptanceAttendanceInterval,
		AcceptanceLeaveInterval:      data.AcceptanceLeaveInterval,
		DefaultYearlyQuota:           data.DefaultYearlyQuota,
		DefaultMarriageQuota:         data.DefaultMarriageQuota,
	}
}

func MapConfigChangesLogToResponse(logs []entity.ConfigurationChangesLog) []dto.ConfigChangesLogsResponse {
	var res []dto.ConfigChangesLogsResponse

	for _, v := range logs {
		res = append(res, dto.ConfigChangesLogsResponse{
			ID: v.ID,
			UpdatedBy: dto.BriefEmployeeListResponse{
				ID:       v.UpdatedByID,
				FullName: v.UpdatedBy.FullName,
				Email:    v.UpdatedBy.Email,
				Avatar:   v.UpdatedBy.Avatar,
				Job:      v.UpdatedBy.Job.Name,
			},
			Changes:     v.Changes,
			UpdatedAt:   v.UpdatedAt.In(utils.CURRENT_LOC).Format(time.RFC1123),
			WhenApplied: v.WhenApplied.In(utils.CURRENT_LOC).Format(time.RFC1123),
		})
	}

	return res
}

/*
*************************************************
REQUEST TO ENTITIES
*************************************************
*/
func MapUpdateConfigToDomain(req dto.UpdateConfigRequest) entity.Configuration {
	return entity.Configuration{
		OfficeStartTimeHour:          req.StartTimeHour,
		OfficeStartTimeMinute:        req.StartTimeMinute,
		OfficeEndTimeHour:            req.EndTimeHour,
		OfficeEndTimeMinute:          req.EndTimeMinute,
		AcceptanceAttendanceInterval: req.AcceptanceAttendanceInterval,
		AcceptanceLeaveInterval:      req.AcceptanceLeaveInterval,
		DefaultYearlyQuota:           req.DefaultYearlyQuota,
		DefaultMarriageQuota:         req.DefaultMarriageQuota,
	}
}
