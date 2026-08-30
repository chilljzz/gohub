package controller

import (
	"log"
	"net/http"

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
	manager *ws.Manager
	handler ws.MessageHandler
}

func NewWSController(
	manager *ws.Manager,
	handler ws.MessageHandler,
) *WSController {
	return &WSController{
		manager: manager,
		handler: handler,
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
	)
	c.manager.Register(client)

	go client.WriteLoop()

	client.ReadLoop()

}
