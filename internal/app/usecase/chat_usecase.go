package usecase

import (
	"context"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
)

type chatUseCase struct {
	chatRepo  repo.IChatRepo
	emplRepo  repo.IEmployeeRepo
	psService service.IPubSubService
}

func NewChatUseCase(chatRepo repo.IChatRepo, emplRepo repo.IEmployeeRepo, psService service.IPubSubService) *chatUseCase {
	return &chatUseCase{
		chatRepo:  chatRepo,
		emplRepo:  emplRepo,
		psService: psService,
	}
}

func (uc *chatUseCase) OpenChat(ctx context.Context, user entity.Employee, room entity.Room) (entity.Room, []entity.Chat, error) {
	room, err := uc.chatRepo.FindOrCreateRoom(ctx, room)
	if err != nil {
		return room, nil, err
	}

	chats, err := uc.chatRepo.GetChatsByRoomId(ctx, room.ID, user.ID)
	if err != nil {
		return room, nil, NewRepositoryError("Chat", err)
	}

	return room, chats, nil
}

func (uc *chatUseCase) SendMessage(ctx context.Context, userId, roomId, message string) (entity.Chat, error) {
	if _, err := uc.chatRepo.FindRoomByID(ctx, roomId); err != nil {
		return entity.Chat{}, NewClientError("Room", err)
	}

	chat, err := uc.chatRepo.CreateNewMessage(ctx, userId, roomId, message)
	if err != nil {
		return chat, NewRepositoryError("Chat", err)
	}

	if err := uc.psService.PublishChat(ctx, roomId, userId, chat); err != nil {
		return chat, NewServiceError("Chat", err)
	}

	return chat, nil
}

func (uc *chatUseCase) ListenMessage(ctx context.Context, userId, roomId string, channel chan entity.Chat) error {
	if err := uc.psService.SubscribeChat(ctx, roomId, userId, channel); err != nil {
		return NewServiceError("Chat", err)
	}

	return nil
}

func (uc *chatUseCase) DetachListener(ctx context.Context, userId, roomId string) error {
	if err := uc.psService.UnregisterClient(ctx, roomId, userId); err != nil {
		return NewServiceError("Chat", err)
	}

	return nil
}
