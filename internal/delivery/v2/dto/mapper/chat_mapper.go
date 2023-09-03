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
			ID:           room.ID.Hex(),
			Participants: room.Participants,
			CreatedAt:    room.CreatedAt.Time().In(utils.CURRENT_LOC).Format(time.RFC1123),
		},
		Chats: []dto.ChatResponse{},
	}

	for _, v := range chats {
		res.Chats = append(res.Chats, MapChatDomainToResponse(v))
	}

	return res
}

func MapOpenChatRoomRequestToDomain(req vo.OpenChatRequest) entity.Room {
	return entity.Room{
		Participants: []string{req.SenderId, req.RecipientId},
	}
}

func MapChatDomainToResponse(chat entity.Chat) dto.ChatResponse {
	return dto.ChatResponse{
		ID:        chat.ID.Hex(),
		RoomId:    chat.RoomId.Hex(),
		Sender:    chat.Sender,
		Message:   chat.Message,
		Read:      chat.Read,
		SentAt:    time.Unix(int64(chat.Timestamp.T), 0).In(utils.CURRENT_LOC).Format(time.RFC1123),
		Timestamp: chat.Timestamp.T,
	}
}

func MapFriendsList(friends []entity.Employee, userId string) []dto.BriefEmployeeListResponse {
	for i, v := range friends {
		if v.ID == userId {
			friends = append(friends[:i], friends[i+1:]...)
			break
		}
	}

	return MapEmployeeListToBriefEmployeeListResponse(friends)
}
