package repo

import (
	"context"

	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type IConfigRepo interface {
	GetConfiguration(ctx context.Context) (entity.Configuration, error)
	SaveNextDayChangesAndLogs(ctx context.Context, config entity.Configuration, logs entity.ConfigurationChangesLog) error
	SaveNextMonthChangesAndLogs(ctx context.Context, config entity.Configuration, logs entity.ConfigurationChangesLog) error
	GetConfigChangesLogs(ctc context.Context, q vo.CommonQuery) ([]entity.ConfigurationChangesLog, vo.PaginationDTOResponse, error)
}
