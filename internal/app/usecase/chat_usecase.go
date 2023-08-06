package usecase

import (
	"context"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/entity"
)

type chatUseCase struct {
	chatRepo repo.IChatRepo
	emplRepo repo.IEmployeeRepo
}

func NewChatUseCase(chatRepo repo.IChatRepo, emplRepo repo.IEmployeeRepo) *chatUseCase {
	return &chatUseCase{
		chatRepo: chatRepo,
		emplRepo: emplRepo,
	}
}

func (uc *chatUseCase) OpenChat(ctx context.Context, user entity.Employee, room entity.Room) (entity.Room, []entity.Chat, error) {
	room, err := uc.chatRepo.FindOrCreateRoom(ctx, room)
	if err != nil {
		return room, nil, err
	}

	chats, err := uc.chatRepo.GetChatsByRoomId(ctx, room.Id, user.Id)
	if err != nil {
		return room, nil, NewRepositoryError("Chat", err)
	}

	return room, chats, nil
}

func (uc *chatUseCase) StoreMessage(ctx context.Context, userId, roomId, message string) (entity.Chat, error) {
	if _, err := uc.chatRepo.FindRoomByID(ctx, roomId); err != nil {
		return entity.Chat{}, NewClientError("Room", err)
	}

	chat, err := uc.chatRepo.CreateNewMessage(ctx, userId, roomId, message)
	if err != nil {
		return chat, NewRepositoryError("Chat", err)
	}

	return chat, nil
}
