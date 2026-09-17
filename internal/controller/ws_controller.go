package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSController struct {
	manager  *ws.Manager
	handler  ws.MessageHandler
	presence ws.PresenceTracker
}

func NewWSController(
	manager *ws.Manager,
	handler ws.MessageHandler,
	presence ws.PresenceTracker,
) *WSController {
	return &WSController{
		manager:  manager,
		handler:  handler,
		presence: presence,
	}
}

func (c *WSController) Connect(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		response.Fail(
			ctx,
			response.CodeUnauthorized,
			"user identity not found",
		)
		return
	}

	conn, err := upgrade.Upgrade(
		ctx.Writer,
		ctx.Request,
		nil,
	)
	if err != nil {
		log.Println("websocket upgrade failed:", err)
		return
	}

	client := ws.NewClient(
		userID,
		conn,
		c.manager,
		c.handler,
		c.presence,
	)
	c.manager.Register(client)

	touchCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := c.presence.Touch(touchCtx, userID); err != nil {
		log.Printf("initial presence touch failed: user_id=%d err=%v", userID, err)
	}
	cancel()

	go client.WriteLoop()

	client.ReadLoop()

}
