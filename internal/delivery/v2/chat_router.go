package v2

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sinarlog.com/internal/app/usecase"
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
}

// Websocket Upgrader
var upgrader = websocket.Upgrader{
	HandshakeTimeout:  5 * time.Second,
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: false,
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		w.WriteHeader(status)
		fmt.Fprintf(w, "error: %s", reason)
	},
}

type Chatter struct {
	roomId     string
	chatterId  string
	conn       *websocket.Conn
	mu         sync.Mutex
	controller *ChatController
}

func NewChatController(rg *gin.RouterGroup, chatUC usecase.IChatUseCase) {
	controller := new(ChatController)
	controller.chatUC = chatUC

	// Normal HTTP
	rg.PUT("/room", controller.openRoomChatHandler)
	// Websockets
	rg.GET("/messenger/:userId/:roomId", controller.chattingHandler)
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

func (controller *ChatController) chattingHandler(c *gin.Context) {
	// Checks whether the connection is websocket
	if !c.IsWebsocket() {
		controller.ClientError(c, usecase.NewClientError("Chat", fmt.Errorf("only websocket connection is allowed")))
		return
	}

	// Gets the user id and room id
	userId := c.Param("userId")
	roomId := c.Param("roomId")

	// Upgrade to websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		controller.UnexpectedError(c, usecase.NewServiceError("Chat", fmt.Errorf("unable to upgrade connection")))
		return
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
