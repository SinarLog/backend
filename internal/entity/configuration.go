package entity

import (
	"time"

	"sinarlog.com/internal/utils"
)

type RedisConfigKeyName string

const (
	CONFIG_KEY_NEXT_DAY   = "configNextDay"
	CONFIG_KEY_NEXT_MONTH = "configNextMonth"
)

type Configuration struct {
	BaseModelId

	OfficeStartTime              time.Time
	OfficeEndTime                time.Time
	OfficeStartTimeHour          int    `gorm:"-:all" redis:"officeStartTimeHour"`
	OfficeStartTimeMinute        int    `gorm:"-:all" redis:"officeStartTimeMinute"`
	OfficeEndTimeHour            int    `gorm:"-:all" redis:"officeEndTimeHour"`
	OfficeEndTimeMinute          int    `gorm:"-:all" redis:"officeEndTimeMinute"`
	AcceptanceAttendanceInterval string `gorm:"type:varchar(50)" redis:"acceptanceAttendanceInterval"` // e.g. 30m, 1h, ...
	AcceptanceLeaveInterval      int    `gorm:"default:7" redis:"acceptanceLeaveInterval"`             // days
	DefaultYearlyQuota           int    `gorm:"default:12" redis:"defaultYearlyQuota"`                 // days
	DefaultMarriageQuota         int    `gorm:"default:3" redis:"defaultMarriageQuota"`                // days
	MaxOvertimeDailyDur          int    `gorm:"default:3" redis:"-"`                                   // hours
	MaxOvertimeWeeklyDur         int    `gorm:"default:14" redis:"-"`                                  // hours

	ConfigurationChangesLogs []ConfigurationChangesLog

	BaseModelStamps
	BaseModelSoftDelete
}

type ConfigurationChangesLog struct {
	BaseModelId

	ConfigurationID string `gorm:"type:uuid"`
	Configuration   Configuration
	UpdatedByID     string `gorm:"type:uuid"`
	UpdatedBy       Employee
	Changes         JSONB
	WhenApplied     time.Time

	BaseModelStamps
	BaseModelSoftDelete
}

// Returns the office work duration that is office end time - office start time
func (c Configuration) OfficeWorkDuration() time.Duration {
	now := time.Now().In(utils.CURRENT_LOC)
	officeStartTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		c.OfficeStartTime.Hour(),
		c.OfficeStartTime.Minute(),
		c.OfficeStartTime.Second(),
		c.OfficeStartTime.Nanosecond(),
		utils.CURRENT_LOC,
	)
	officeEndTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		c.OfficeEndTime.Hour(),
		c.OfficeEndTime.Minute(),
		c.OfficeEndTime.Second(),
		c.OfficeEndTime.Nanosecond(),
		utils.CURRENT_LOC,
	)

	return officeEndTime.Sub(officeStartTime)
}
