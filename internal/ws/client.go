package ws

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte

	manager  *Manager
	handler  MessageHandler
	presence PresenceTracker

	closed    chan struct{}
	closeOnce sync.Once
}

func NewClient(
	userID uint,
	conn *websocket.Conn,
	manager *Manager,
	handler MessageHandler,
	presence PresenceTracker,
) *Client {
	return &Client{
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		manager:  manager,
		handler:  handler,
		presence: presence,
		closed:   make(chan struct{}),
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)

		c.manager.Unregister(c)

		if err := c.Conn.Close(); err != nil {

			slog.Error(
				"user  close websocket error",
				slog.Uint64("user_id", uint64(c.UserID)),
				slog.Any("error", err),
			)

		}

	})
}

func (c *Client) SendMessage(message []byte) bool {

	select {
	case <-c.closed:
		return false
	default:
	}

	select {
	case c.Send <- message:
		return true
	case <-c.closed:
		return false
	default:
		return false
	}

}

func (c *Client) ReadLoop(ctx context.Context) {
	defer c.Close()
	defer c.recoverLoop("read_loop")

	c.Conn.SetReadLimit(maxMessageSize)

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {

		slog.Error(
			"user set read deadline error",
			slog.Uint64("user_id", uint64(c.UserID)),
			slog.Any("error", err),
		)

		return
	}

	c.Conn.SetPongHandler(
		func(string) error {

			if c.presence != nil {
				ctx, cancel := context.WithTimeout(ctx, time.Second)
				err := c.presence.Touch(ctx, c.UserID)
				cancel()
				if err != nil {
					slog.Warn(
						"presence touch failed",
						slog.Uint64("user_id", uint64(c.UserID)),
						slog.Any("error", err),
					)

				}
			}

			slog.Debug(
				"websocket pong received",

				slog.Uint64("user_id", uint64(c.UserID)),
			)
			return c.Conn.SetReadDeadline(
				time.Now().Add(pongWait),
			)
		},
	)

	for {
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			slog.Error(
				"user read error",
				slog.Uint64("user_id", uint64(c.UserID)),
				slog.Any("error", err),
			)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		c.handler.Handle(ctx, c, data)

	}
}

func (c *Client) WriteLoop(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.Close()
	defer c.recoverLoop("write_loop")
	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-c.Send:
			if !ok {
				return
			}
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error(
					"user set write deadline error",
					slog.Uint64("user_id", uint64(c.UserID)),
					slog.Any("error", err),
				)

				return
			}
			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {

				slog.Error(
					"user  write error",
					slog.Uint64("user_id", uint64(c.UserID)),
					slog.Any("error", err),
				)

				return
			}
		case <-ticker.C:

			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {

				slog.Error(
					"user set ping deadline error",
					slog.Uint64("user_id", uint64(c.UserID)),
					slog.Any("error", err),
				)

				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error(
					"user ping error",
					slog.Uint64("user_id", uint64(c.UserID)),
					slog.Any("error", err),
				)

				return
			}
			slog.Debug(
				"websocket ping sent",

				slog.Uint64("user_id", uint64(c.UserID)),
			)

		case <-c.closed:
			return
		}

	}
}

func (c *Client) recoverLoop(name string) {
	if err := recover(); err != nil {
		slog.Error(
			"websocket panic",

			slog.String("loop", name),

			slog.Uint64("user_id", uint64(c.UserID)),

			slog.Any("error", err),

			slog.String("stack", string(debug.Stack())),
		)
	}
}
