package ws

import (
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte

	manager *Manager
	handler MessageHandler

	closed    chan struct{}
	closeOnce sync.Once
}

func NewClient(
	userID uint,
	conn *websocket.Conn,
	manager *Manager,
	handler MessageHandler,
) *Client {
	return &Client{
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		manager: manager,
		handler: handler,
		closed:  make(chan struct{}),
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)

		c.manager.Unregister(c)

		if err := c.Conn.Close(); err != nil {
			log.Printf(
				"user %d close websocket error:%v\n",
				c.UserID,
				err,
			)
		}

	})
}

func (c *Client) SendMessage(message []byte) bool {
	select {
	case c.Send <- message:
		return true
	case <-c.closed:
		return false
	default:
		return false
	}

}

func (c *Client) ReadLoop() {
	defer c.Close()
	defer c.recoverLoop("read_loop")

	c.Conn.SetReadLimit(maxMessageSize)

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf(
			"user %d set read deadline error: %v\n",
			c.UserID,
			err,
		)
		return
	}

	c.Conn.SetPongHandler(
		func(string) error {
			log.Printf(
				"user %d pong received\n",
				c.UserID,
			)
			return c.Conn.SetReadDeadline(
				time.Now().Add(pongWait),
			)
		},
	)

	for {
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf(
				"user %d read error: %v\n",
				c.UserID,
				err,
			)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		c.handler.Handle(c, data)

	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.Close()
	defer c.recoverLoop("write_loop")
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf(
					"user %d set write deadline error: %v\n",
					c.UserID,
					err,
				)

				return
			}
			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				log.Printf(
					"user %d write error: %v\n",
					c.UserID,
					err,
				)

				return
			}
		case <-ticker.C:

			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf(
					"user %d set ping deadline error: %v\n",
					c.UserID,
					err,
				)

				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf(
					"user %d ping error: %v\n",
					c.UserID,
					err,
				)

				return
			}
			log.Printf(
				"user %d ping sent\n",
				c.UserID,
			)

		case <-c.closed:
			return
		}

	}
}

func (c *Client) recoverLoop(name string) {
	if err := recover(); err != nil {
		log.Printf(
			"websocket panic: loop = %s user_id=%d err=%v\n%s",
			name,
			c.UserID,
			err,
			debug.Stack(),
		)
	}
}
