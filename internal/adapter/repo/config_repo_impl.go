package repo

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type configRepo struct {
	db   *gorm.DB
	rdis *redis.Client
}

func NewConfigRepo(db *gorm.DB, redis *redis.Client) *configRepo {
	return &configRepo{db: db, rdis: redis}
}

func (repo *configRepo) GetConfiguration(ctx context.Context) (entity.Configuration, error) {
	var config entity.Configuration

	if err := repo.db.WithContext(ctx).
		Model(&config).
		First(&config).Error; err != nil {
		return config, err
	}

	config.OfficeStartTime = config.OfficeStartTime.In(utils.CURRENT_LOC)
	config.OfficeEndTime = config.OfficeEndTime.In(utils.CURRENT_LOC)

	return config, nil
}

func (repo *configRepo) SaveNextDayChangesAndLogs(ctx context.Context, config entity.Configuration, logs entity.ConfigurationChangesLog) error {
	tx := repo.db.WithContext(ctx).Begin()

	if err := tx.Model(&logs).Save(&logs).Error; err != nil {
		tx.Rollback()
		return err
	}

	if _, err := repo.rdis.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_DAY, "officeStartTimeHour", config.OfficeStartTimeHour)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_DAY, "officeStartTimeMinute", config.OfficeStartTimeMinute)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_DAY, "officeEndTimeHour", config.OfficeEndTimeHour)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_DAY, "officeEndTimeMinute", config.OfficeEndTimeMinute)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_DAY, "acceptanceAttendanceInterval", config.AcceptanceAttendanceInterval)
		return nil
	}); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (repo *configRepo) SaveNextMonthChangesAndLogs(ctx context.Context, config entity.Configuration, logs entity.ConfigurationChangesLog) error {
	tx := repo.db.WithContext(ctx).Begin()

	if err := tx.Model(&logs).Save(&logs).Error; err != nil {
		tx.Rollback()
		return err
	}

	if _, err := repo.rdis.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_MONTH, "acceptanceLeaveInterval", config.AcceptanceLeaveInterval)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_MONTH, "defaultYearlyQuota", config.DefaultYearlyQuota)
		p.HSet(ctx, entity.CONFIG_KEY_NEXT_MONTH, "defaultMarriageQuota", config.DefaultMarriageQuota)
		return nil
	}); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (repo *configRepo) GetConfigChangesLogs(ctx context.Context, q vo.CommonQuery) ([]entity.ConfigurationChangesLog, vo.PaginationDTOResponse, error) {
	pquery := q.Pagination.MustExtract()

	var changes []entity.ConfigurationChangesLog
	var count int64

	if err := repo.db.WithContext(ctx).
		Model(&entity.ConfigurationChangesLog{}).
		Preload("UpdatedBy.Job").
		Count(&count).
		Order(utils.ToOrderSQL(pquery.OrderBy, pquery.Sort)).
		Limit(pquery.Limit).
		Offset(pquery.Offset).
		Find(&changes).Error; err != nil {
		return nil, vo.PaginationDTOResponse{}, err
	}

	return changes, pquery.Compress(count), nil
}
