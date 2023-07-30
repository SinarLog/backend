package service

import (
	"context"
	"time"

	"sinarlog.com/internal/entity"
)

type IDoorkeeperService interface {
	HashPassword(pass string) ([]byte, error)
	VerifyPassword(hash, password string) error
	VerifyAndParseToken(ctx context.Context, tk string) (string, error)

	GenerateOTP() (string, int64, time.Duration)
	VerifyOTP(otp string, timestamp int64) bool

	// V2
	GenerateToken(employee entity.Employee) (string, error)
}
