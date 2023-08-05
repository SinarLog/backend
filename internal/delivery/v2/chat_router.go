package v2

import (
	"github.com/gin-gonic/gin"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type ChatController struct {
	model.BaseControllerV2
	chatUC usecase.IChatUseCase
}

func NewChatController(rg *gin.RouterGroup, chatUC usecase.IChatUseCase) {
	controller := new(ChatController)
	controller.chatUC = chatUC

	rg.PUT("/room", controller.openRoomChatHandler)
}

func (controller *ChatController) openRoomChatHandler(c *gin.Context) {
	var req vo.OpenChatRequest

	user := c.Keys["user"].(entity.Employee)

	if err := c.ShouldBindJSON(&req); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body", err))
		return
	}

	room, chats, err := controller.chatUC.OpenChat(c.Request.Context(), user, mapper.MapOpenChatRoomRequestToDomain(req))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapOpenChatResponse(room, chats))
}
