package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	Rooms string = "rooms"
	Chats string = "chats"
)

type Room struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	Participants []string           `bson:"participants,omitempty"`
	CreatedAt    primitive.DateTime `bson:"createdAt,omitempty"`
}

type Chat struct {
	Id        primitive.ObjectID  `bson:"_id,omitempty"`
	RoomId    primitive.ObjectID  `bson:"roomId,omitempty"`
	Sender    string              `bson:"sender,omitempty"`
	Message   string              `bson:"message,omitempty"`
	Read      bool                `bson:"read"`
	Timestamp primitive.Timestamp `bson:"timestamp,omitempty"`
}
