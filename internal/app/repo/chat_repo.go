package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"sinarlog.com/internal/entity"
)

type IChatRepo interface {
	FindOrCreateRoom(ctx context.Context, room entity.Room) (entity.Room, error)
	CreateRoom(ctx context.Context, room entity.Room) (entity.Room, error)
	FindRoom(ctx context.Context, room entity.Room) (entity.Room, error)
	FindRoomByID(ctx context.Context, id string) (entity.Room, error)
	GetChatsByRoomId(ctx context.Context, roomId primitive.ObjectID, readerId string) ([]entity.Chat, error)
}
