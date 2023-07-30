package service

import (
	"context"
	"mime/multipart"
)

type IBucketService interface {
	CreateAvatar(ctx context.Context, employeeId string, file multipart.File) (string, error)
	CreateLeaveAttachment(ctx context.Context, leaveId string, file multipart.File) (string, error)
	DeleteAvatar(ctx context.Context, employeeId string) error
	DeleteLeaveAttachment(ctx context.Context, leaveId string) error
}
