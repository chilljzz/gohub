package controller

import (
	"context"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(
		r *http.Request,
	) bool {

		origin := r.Header.Get(
			"Origin",
		)

		if origin == "" {
			// curl/Postman 等非浏览器客户端
			return true
		}

		parsedOrigin, err := url.Parse(
			origin,
		)

		if err != nil {
			return false
		}

		return strings.EqualFold(
			parsedOrigin.Host,
			r.Host,
		)
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

	connCtx, cancel :=
		context.WithCancel(
			ctx.Request.Context(),
		)
	defer cancel()

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
	if ok := c.manager.Register(client); !ok {
		client.Close()
		return
	}

	touchCtx, cancel := context.WithTimeout(connCtx, time.Second)
	if err := c.presence.Touch(touchCtx, userID); err != nil {
		log.Printf("initial presence touch failed: user_id=%d err=%v", userID, err)
	}
	cancel()

	go client.WriteLoop(connCtx)

	client.ReadLoop(connCtx)

}
