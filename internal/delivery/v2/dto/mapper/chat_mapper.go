package mapper

import (
	"time"

	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

func MapOpenChatResponse(room entity.Room, chats []entity.Chat) dto.OpenChatResponse {
	res := dto.OpenChatResponse{
		Room: dto.RoomResponse{
			Id:           room.Id.Hex(),
			Participants: room.Participants,
			CreatedAt:    room.CreatedAt.Time().In(utils.CURRENT_LOC).Format(time.RFC1123),
		},
		Chat: []dto.ChatResponse{},
	}

	for _, v := range chats {
		res.Chat = append(res.Chat, dto.ChatResponse{
			Id:        v.Id.Hex(),
			RoomId:    v.RoomId.Hex(),
			Sender:    v.Sender,
			Message:   v.Message,
			Read:      v.Read,
			SentAt:    time.Unix(int64(v.Timestamp.T), 0).In(utils.CURRENT_LOC).Format(time.RFC1123),
			Timestamp: v.Timestamp.I,
		})
	}

	return res
}

func MapOpenChatRoomRequestToDomain(req vo.OpenChatRequest) entity.Room {
	return entity.Room{
		Participants: []string{req.SenderId, req.RecipientId},
	}
}
