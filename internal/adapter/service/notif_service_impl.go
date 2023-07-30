package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"sinarlog.com/internal/entity"
)

var (
	baseChannel string = "app:notif"
	overtime    string = "overtime"
	leave       string = "leave"
)

type notifService struct {
	rdis *redis.Client
}

func NewNotifService(redis *redis.Client) *notifService {
	return &notifService{redis}
}

func (s *notifService) SendOvertimeSubmissionNotification(ctx context.Context, receiver, sender entity.Employee) (int64, error) {
	channel := fmt.Sprintf("%s:%s", baseChannel, receiver.Id)
	payload := fmt.Sprintf("%s;%s;%s", overtime, sender.FullName, sender.Avatar)
	res, err := s.rdis.Publish(ctx, channel, payload).Result()
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (s *notifService) SendLeaveRequestNotification(ctx context.Context, receiver, sender entity.Employee) (int64, error) {
	channel := fmt.Sprintf("%s:%s", baseChannel, receiver.Id)
	payload := fmt.Sprintf("%s;%s;%s", leave, sender.FullName, sender.Avatar)
	res, err := s.rdis.Publish(ctx, channel, payload).Result()
	if err != nil {
		return 0, err
	}

	return res, nil
}
