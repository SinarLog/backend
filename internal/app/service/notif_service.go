package service

import (
	"context"

	"sinarlog.com/internal/entity"
)

type INotifService interface {
	SendOvertimeSubmissionNotification(ctx context.Context, receiver, sender entity.Employee) (int64, error)
	SendLeaveRequestNotification(ctx context.Context, receiver, sender entity.Employee) (int64, error)
}
