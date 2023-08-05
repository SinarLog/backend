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
	room, err := uc.openOrCreateRoom(ctx, room)
	if err != nil {
		return room, nil, err
	}

	chats, err := uc.chatRepo.GetChatsByRoomId(ctx, room.Id, user.Id)
	if err != nil {
		return room, nil, NewRepositoryError("Chat", err)
	}

	return room, chats, nil
}

func (uc *chatUseCase) openOrCreateRoom(ctx context.Context, room entity.Room) (entity.Room, error) {
	room, err := uc.chatRepo.FindRoom(ctx, room)
	if err != nil {
		room, err = uc.chatRepo.CreateRoom(ctx, room)
		if err != nil {
			return room, NewRepositoryError("Room", err)
		}
	}

	return room, nil
}
