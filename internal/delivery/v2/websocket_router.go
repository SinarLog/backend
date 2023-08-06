package v2

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/pkg/redis"
)

type WebsocketController struct {
	model.BaseControllerV2
}

func NewWebsocketController(rg *gin.RouterGroup) {
	controller := new(WebsocketController)

	rg.GET(":id", controller.connectionHandlers)
}

func (controller *WebsocketController) connectionHandlers(c *gin.Context) {
	if !c.IsWebsocket() {
		controller.ClientError(c, usecase.NewClientError("Notification", fmt.Errorf("only websocket connection is allowed")))
		return
	}

	user := c.Param("id")
	conn, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		controller.UnexpectedError(c, usecase.NewServiceError("Notification", fmt.Errorf("unable to upgrade connection")))
	}

	// Make a goroutine to handle close connections
	// by sending signal to a channel
	connChan := make(chan bool)
	go func(ch chan<- bool, conn *websocket.Conn) {
		_, _, err := conn.ReadMessage()
		if ce, ok := err.(*websocket.CloseError); ok {
			switch ce.Code {
			case websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived:
				connChan <- true
				return
			}
		}
	}(connChan, conn)

	// Start subscribing to redis pubsub
	rdis := redis.NewRedisClient()
	channel := fmt.Sprintf("%s:%s", "app:notif", user)
	pubsub := rdis.Client.Subscribe(c.Request.Context(), channel)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-connChan:
			// This signals that the client disconnected
			if err := conn.Close(); err != nil {
				log.Fatalf("Unable to close websocket connection: %s\n", err)
			}
			return
		case msg := <-ch:
			// This signals that there is a message from redis publishers
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				log.Fatalf("Unable to send meassage: %s\n", err.Error())
				return
			}
		}
	}
}
