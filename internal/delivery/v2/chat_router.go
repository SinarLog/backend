package v2

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/middleware"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

/*
Controller types
*/
type ChatController struct {
	model.BaseControllerV2
	chatUC usecase.IChatUseCase
	emplUC usecase.IEmployeeUseCase
}

type Chatter struct {
	roomId     string
	chatterId  string
	conn       *websocket.Conn
	mu         sync.Mutex
	controller *ChatController
}

func NewChatController(rg *gin.RouterGroup, credUC usecase.ICredentialUseCase, emplUC usecase.IEmployeeUseCase, chatUC usecase.IChatUseCase) {
	controller := new(ChatController)
	controller.chatUC = chatUC
	controller.emplUC = emplUC

	// Normal HTTP
	rg.GET("/friends", middleware.NewMiddleware().AuthMiddleware(credUC, "hr", "mngr", "staff"), controller.getFriendsHandler)
	rg.PUT("/room", middleware.NewMiddleware().AuthMiddleware(credUC, "hr", "mngr", "staff"), controller.openRoomChatHandler)
	// Websockets
	rg.GET("/messenger/:roomId/:userId", controller.chattingHandler)
}

func (controller *ChatController) getFriendsHandler(c *gin.Context) {
	pquery := controller.ParsePagination(c)
	user := c.Keys["user"].(entity.Employee)

	q := vo.AllEmployeeQuery{
		CommonQuery: vo.CommonQuery{Pagination: pquery},
	}

	res, _, err := controller.emplUC.RetrieveEmployeesList(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapFriendsList(res, user.Id))
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

	controller.Ok(c, mapper.MapOpenChatResponse(room, chats))
}

func (controller *ChatController) chattingHandler(c *gin.Context) {
	// Checks whether the connection is websocket
	if !c.IsWebsocket() {
		controller.ClientError(c, usecase.NewClientError("Chat", fmt.Errorf("only websocket connection is allowed")))
		return
	}

	// Gets the user id and room id
	userId := c.Param("userId")
	roomId := c.Param("roomId")

	conn, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		controller.UnexpectedError(c, usecase.NewServiceError("Notification", fmt.Errorf("unable to upgrade connection")))
	}

	chatter := &Chatter{
		roomId:     roomId,
		chatterId:  userId,
		conn:       conn,
		controller: controller,
	}

	closed := make(chan int)
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Sender
	go chatter.messageSender(ctx, closed)
	// Reader
	go chatter.messageListener(ctx)

	<-closed
}

/*
***************************
SOCKET MANAGEMENT HELPERS
***************************
*/
func (client *Chatter) messageSender(ctx context.Context, closeChan chan int) {
	for {
		_, message, err := client.conn.ReadMessage()
		if ce, ok := err.(*websocket.CloseError); ok {
			switch ce.Code {
			case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived:
				// Close conn
				client.CloseConn()
				// Unregister client
				client.controller.chatUC.DetachListener(ctx, client.chatterId, client.roomId)
				// Let the subscriber goroutine know
				closeChan <- 1
				return
			}
		}

		_, err = client.controller.chatUC.SendMessage(ctx, client.chatterId, client.roomId, string(message))
		if err != nil {
			client.conn.WriteJSON(err)
		}
	}
}

func (client *Chatter) messageListener(ctx context.Context) {
	channel := make(chan entity.Chat)

	go func(ctx context.Context, channel chan entity.Chat) {
		if err := client.controller.chatUC.ListenMessage(ctx, client.chatterId, client.roomId, channel); err != nil {
			client.conn.WriteJSON(err)
			return
		}
	}(ctx, channel)

	for {
		select {
		case <-ctx.Done():
			return
		case chat := <-channel:
			client.conn.WriteJSON(mapper.MapChatDomainToResponse(chat))
		}
	}
}

func (client *Chatter) CloseConn() {
	client.mu.Lock()
	defer client.mu.Unlock()

	if err := client.conn.Close(); err != nil {
		return
	}
}
