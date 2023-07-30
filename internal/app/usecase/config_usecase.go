package usecase

import (
	"context"
	"fmt"
	"time"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type configUseCase struct {
	configRepo repo.IConfigRepo
}

func NewConfigUseCase(configRepo repo.IConfigRepo) *configUseCase {
	return &configUseCase{configRepo}
}

func (uc *configUseCase) RetrieveConfiguration(ctx context.Context) (entity.Configuration, error) {
	config, err := uc.configRepo.GetConfiguration(ctx)
	if err != nil {
		return entity.Configuration{}, NewRepositoryError("Config", err)
	}

	config.OfficeStartTime = config.OfficeStartTime.In(utils.CURRENT_LOC)
	config.OfficeEndTime = config.OfficeEndTime.In(utils.CURRENT_LOC)

	return config, err
}

func (uc *configUseCase) RetrieveChangesLogs(ctx context.Context, q vo.CommonQuery) ([]entity.ConfigurationChangesLog, vo.PaginationDTOResponse, error) {
	changes, page, err := uc.configRepo.GetConfigChangesLogs(ctx, q)
	if err != nil {
		return changes, page, NewRepositoryError("Config", err)
	}

	return changes, page, nil
}

func (uc *configUseCase) ChangeCompanyConfig(ctx context.Context, hr entity.Employee, payload entity.Configuration) error {
	config, err := uc.configRepo.GetConfiguration(ctx)
	if err != nil {
		return NewRepositoryError("Config", err)
	}

	if err := uc.changeCompanyConfigNextDay(ctx, hr, config, payload); err != nil {
		return err
	}

	if err := uc.changeCompanyConfigNextMonth(ctx, hr, config, payload); err != nil {
		return err
	}

	return nil
}

func (uc *configUseCase) changeCompanyConfigNextDay(ctx context.Context, hr entity.Employee, config, payload entity.Configuration) error {
	var changes map[string]any = make(map[string]any)
	var logs entity.ConfigurationChangesLog

	now := time.Now().In(utils.CURRENT_LOC)

	logs.Configuration = config
	logs.ConfigurationID = config.Id
	logs.UpdatedBy = hr
	logs.UpdatedByID = hr.Id
	logs.WhenApplied = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, utils.CURRENT_LOC)

	// Start time
	if config.OfficeStartTime.Hour() != payload.OfficeStartTimeHour || config.OfficeStartTime.Minute() != payload.OfficeStartTimeMinute {
		change := make(map[string]string)
		change["prev"] = fmt.Sprintf("%s:%s", utils.PadIntegerTime(config.OfficeStartTime.Hour()), utils.PadIntegerTime(config.OfficeStartTime.Minute()))
		change["new"] = fmt.Sprintf("%s:%s", utils.PadIntegerTime(payload.OfficeStartTimeHour), utils.PadIntegerTime(payload.OfficeStartTimeMinute))
		changes["office_start_time"] = any(change)
	}
	config.OfficeStartTimeHour = payload.OfficeStartTimeHour
	config.OfficeStartTimeMinute = payload.OfficeStartTimeMinute

	// End time
	if config.OfficeEndTime.Hour() != payload.OfficeEndTimeHour || config.OfficeEndTime.Minute() != payload.OfficeEndTimeMinute {
		change := make(map[string]string)
		change["prev"] = fmt.Sprintf("%s:%s", utils.PadIntegerTime(config.OfficeEndTime.Hour()), utils.PadIntegerTime(config.OfficeEndTime.Minute()))
		change["new"] = fmt.Sprintf("%s:%s", utils.PadIntegerTime(payload.OfficeEndTimeHour), utils.PadIntegerTime(payload.OfficeEndTimeMinute))
		changes["office_end_time"] = any(change)
	}
	config.OfficeEndTimeHour = payload.OfficeEndTimeHour
	config.OfficeEndTimeMinute = payload.OfficeEndTimeMinute

	// Acceptance attendance interval
	if config.AcceptanceAttendanceInterval != payload.AcceptanceAttendanceInterval {
		change := make(map[string]string)
		change["prev"] = config.AcceptanceAttendanceInterval
		change["new"] = payload.AcceptanceAttendanceInterval
		changes["acceptance_attendance_interval"] = any(change)
	}
	config.AcceptanceAttendanceInterval = payload.AcceptanceAttendanceInterval

	if len(changes) != 0 {
		logs.Changes = changes
		if err := uc.configRepo.SaveNextDayChangesAndLogs(ctx, config, logs); err != nil {
			return NewRepositoryError("Config", err)
		}
	}

	return nil
}

func (uc *configUseCase) changeCompanyConfigNextMonth(ctx context.Context, hr entity.Employee, config, payload entity.Configuration) error {
	var changes map[string]any = make(map[string]any)
	var logs entity.ConfigurationChangesLog

	now := time.Now().In(utils.CURRENT_LOC)

	logs.Configuration = config
	logs.ConfigurationID = config.Id
	logs.UpdatedBy = hr
	logs.UpdatedByID = hr.Id
	logs.WhenApplied = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, utils.CURRENT_LOC)

	if config.AcceptanceLeaveInterval != payload.AcceptanceLeaveInterval {
		change := make(map[string]int)
		change["prev"] = config.AcceptanceLeaveInterval
		change["new"] = payload.AcceptanceLeaveInterval
		changes["acceptance_leave_interval"] = any(change)
	}
	config.AcceptanceLeaveInterval = payload.AcceptanceLeaveInterval

	if config.DefaultYearlyQuota != payload.DefaultYearlyQuota {
		change := make(map[string]int)
		change["prev"] = config.DefaultYearlyQuota
		change["new"] = payload.DefaultYearlyQuota
		changes["default_yearly_quota"] = any(change)
	}
	config.DefaultYearlyQuota = payload.DefaultYearlyQuota

	if config.DefaultMarriageQuota != payload.DefaultMarriageQuota {
		change := make(map[string]int)
		change["prev"] = config.DefaultMarriageQuota
		change["new"] = payload.DefaultMarriageQuota
		changes["default_marriage_quota"] = any(change)
	}
	config.DefaultMarriageQuota = payload.DefaultMarriageQuota

	if len(changes) != 0 {
		logs.Changes = changes
		if err := uc.configRepo.SaveNextMonthChangesAndLogs(ctx, config, logs); err != nil {
			return NewRepositoryError("Config", err)
		}
	}

	return nil
}
