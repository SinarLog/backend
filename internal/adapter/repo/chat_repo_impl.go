package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

type chatRepo struct {
	chatColl *mongo.Collection
	roomColl *mongo.Collection
}

func NewChatRepo(mgDB *mongo.Database) *chatRepo {
	return &chatRepo{
		roomColl: mgDB.Collection(entity.Rooms),
		chatColl: mgDB.Collection(entity.Chats),
	}
}

func (repo *chatRepo) FindOrCreateRoom(ctx context.Context, room entity.Room) (entity.Room, error) {
	room, err := repo.FindRoom(ctx, room)
	if err != nil {
		room, err = repo.CreateRoom(ctx, room)
		if err != nil {
			return room, err
		}
	}

	return room, nil
}

func (repo *chatRepo) FindRoom(ctx context.Context, room entity.Room) (entity.Room, error) {
	if err := repo.roomColl.FindOne(ctx, bson.D{
		{Key: "participants", Value: bson.D{
			{Key: "$all", Value: room.Participants},
		}},
	}).Decode(&room); err != nil {
		return room, err
	}

	return room, nil
}

func (repo *chatRepo) FindRoomByID(ctx context.Context, id string) (entity.Room, error) {
	var room entity.Room

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return room, err
	}

	if err := repo.roomColl.FindOne(ctx, bson.M{"_id": _id}).Decode(&room); err != nil {
		return room, err
	}

	return room, nil
}

func (repo *chatRepo) CreateRoom(ctx context.Context, room entity.Room) (entity.Room, error) {
	res, err := repo.roomColl.InsertOne(ctx, room, options.InsertOne().SetComment("a new room has been created"))
	if err != nil {
		return room, err
	}

	room.ID = res.InsertedID.(primitive.ObjectID)

	return room, nil
}

func (repo *chatRepo) GetChatsByRoomId(ctx context.Context, roomId primitive.ObjectID, readerId string) ([]entity.Chat, error) {
	var chats []entity.Chat

	cursor, err := repo.chatColl.Find(ctx,
		bson.D{{Key: "roomId", Value: roomId}},
		options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}),
		options.Find().SetLimit(30),
	)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var chat entity.Chat
		if err := cursor.Decode(&chat); err != nil {
			return nil, err
		}
		chat.Read = true
		chats = append(chats, chat)
	}

	if _, err := repo.chatColl.UpdateMany(ctx,
		bson.D{
			{Key: "roomId", Value: roomId},
			{Key: "sender", Value: bson.D{
				{Key: "$ne", Value: readerId},
			}},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "read", Value: true},
			}},
		},
	); err != nil {
		return nil, err
	}
	return chats, nil
}

func (repo *chatRepo) CreateNewMessage(ctx context.Context, userId, roomId, message string) (entity.Chat, error) {
	_id, err := primitive.ObjectIDFromHex(roomId)
	if err != nil {
		return entity.Chat{}, err
	}

	chat := entity.Chat{
		RoomId:  _id,
		Sender:  userId,
		Message: message,
		Read:    false,
		Timestamp: primitive.Timestamp{
			T: uint32(time.Now().In(utils.CURRENT_LOC).Unix()),
			I: 0,
		},
	}

	res, err := repo.chatColl.InsertOne(ctx, chat)
	if err != nil {
		return chat, err
	}

	chat.ID = res.InsertedID.(primitive.ObjectID)

	return chat, nil
}
