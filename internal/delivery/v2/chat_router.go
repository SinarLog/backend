package v2

import (
	"context"
	"fmt"
	"log"
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
Concurrent messaging types
*/
type Hubs struct {
	mu    sync.RWMutex
	rooms map[string]*Hub
}

type Hub struct {
	id      string
	clients map[string]*MessengerClient
	message chan entity.Chat
}

type MessengerClient struct {
	id         string
	roomId     string
	hub        *Hub
	conn       *websocket.Conn
	mu         sync.Mutex
	controller *ChatController
	message    chan entity.Chat
}

/*
Controller types
*/
type ChatController struct {
	model.BaseControllerV2
	chatUC usecase.IChatUseCase
}

var (
	// Websocket Upgrader
	upgrader = websocket.Upgrader{
		HandshakeTimeout:  5 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			w.WriteHeader(status)
			fmt.Fprintf(w, "error: %s", reason)
		},
	}
	// Store all the hubs
	hubs *Hubs
)

func init() {
	hubs = &Hubs{
		rooms: map[string]*Hub{},
	}

	t := time.NewTicker(10 * time.Second)

	go func() {
		for range t.C {
			log.Printf("CURRENT HUBS STATE\n")
			for _, hub := range hubs.rooms {
				log.Printf("%+v", hub)
			}
		}
	}()
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

	// Create the client
	client := &MessengerClient{
		id:         userId,
		roomId:     roomId,
		conn:       conn,
		controller: controller,
		message:    make(chan entity.Chat),
	}

	// Find or create the hub
	hub := client.FindOrCreateHub()
	client.hub = hub

	// Sender
	go client.sendMessage(c.Request.Context())
	// Reader
	go client.readMessage(c.Request.Context())

	for range c.Done() {
		return
	}
}

/*
***************************
SOCKET MANAGEMENT HELPERS
***************************
*/
func (client *MessengerClient) readMessage(ctx context.Context) {
	for {
		_, message, err := client.conn.ReadMessage()
		if ce, ok := err.(*websocket.CloseError); ok {
			switch ce.Code {
			case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived:
				// Close conn
				client.CloseConn()
				// Unsubscribe to hub
				client.UnsubscribeToHub()
				return
			}
		}

		chat, err := client.controller.chatUC.StoreMessage(ctx, client.id, client.hub.id, string(message))
		if err != nil {
			log.Printf("unable to create chat: %s\n", err)
			return
		}

		client.hub.message <- chat
	}
}

func (client *MessengerClient) sendMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case chat := <-client.message:
			client.mu.Lock()

			if err := client.conn.WriteJSON(mapper.MapChatDomainToResponse(chat)); err != nil {
				log.Printf("unable to write chat json: %s\n", err)
				return
			}

			client.mu.Unlock()
		}
	}
}

func (client *MessengerClient) FindOrCreateHub() *Hub {
	hubs.mu.Lock()
	defer hubs.mu.Unlock()

	hub, found := hubs.rooms[client.roomId]
	if !found {
		hub = &Hub{
			id:      client.roomId,
			clients: map[string]*MessengerClient{},
			message: make(chan entity.Chat, 3),
		}

		hub.clients[client.id] = client
		hubs.rooms[client.roomId] = hub

		// Spawn goroutine for the hub
		go hub.spawnWorker()
	} else {
		hub.clients[client.id] = client
	}

	return hub
}

func (hub *Hub) spawnWorker() {
	log.Printf("HUB WORKER with ID %s has spawned...\n", hub.id)
	for {
		select {
		case chat := <-hub.message:
			for _, client := range hub.clients {
				client.message <- chat
			}
		default:
			if len(hub.clients) == 0 {
				// Destroy hub
				close(hub.message)
				delete(hubs.rooms, hub.id)
				log.Printf("HUB WORKER with ID %s has been stopped...\n", hub.id)
				return
			}
		}
	}
}

func (client *MessengerClient) UnsubscribeToHub() {
	hubs.mu.Lock()
	defer hubs.mu.Unlock()

	hub := hubs.rooms[client.roomId]
	delete(hub.clients, client.id)
}

func (client *MessengerClient) CloseConn() {
	client.mu.Lock()
	defer client.mu.Unlock()

	if err := client.conn.Close(); err != nil {
		log.Fatalf("unable to close conn")
	}
}
