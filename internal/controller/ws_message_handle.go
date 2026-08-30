package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/chilljzz/gohub/internal/realtime"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/chilljzz/gohub/internal/ws"
)

type WSMessageHandler struct {
	manager         *ws.Manager
	realtimeService *service.RealtimeService
	messageService  *service.ChannelMessageService
	broker          *realtime.RedisBroker
}

func NewWSMessageHandler(
	manager *ws.Manager,
	realtimeService *service.RealtimeService,
	messageService *service.ChannelMessageService,
	broker *realtime.RedisBroker,
) *WSMessageHandler {
	return &WSMessageHandler{
		manager:         manager,
		realtimeService: realtimeService,
		messageService:  messageService,
		broker:          broker,
	}
}

func (h *WSMessageHandler) Handle(
	client *ws.Client,
	data []byte,
) {
	var req ws.IncomingMessage

	if err := json.Unmarshal(data, &req); err != nil {
		h.sendError(client, "invalid message format")
	}

	switch req.Type {
	case "join_channel":
		h.handleJoinChannle(client, req)

	case "send_message":
		h.handleSendMessage(client, req)

	case "leave_channel":
		h.handleLeaveChannel(client, req)

	case "debug_panic":
		panic("websocket read loop panic test")

	default:
		h.sendError(client, "unkown message type")
	}

}

func (h *WSMessageHandler) handleJoinChannle(
	client *ws.Client,
	req ws.IncomingMessage,
) {
	if req.ChannelID == 0 {
		h.sendError(client, "channel_id is required")
		return
	}
	err := h.realtimeService.CheckChannelAccess(
		client.UserID,
		req.ChannelID,
	)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}
	h.manager.JoinChannel(req.ChannelID, client)

	h.send(client, ws.OutgoingMessage{
		Type:      "joined_channel",
		ChannelID: req.ChannelID,
		Message:   "joined channel successfully",
	})
}

func (h *WSMessageHandler) handleSendMessage(
	client *ws.Client,
	req ws.IncomingMessage,
) {

	if req.ChannelID == 0 {
		h.sendError(client, "channel_id is required")
		return
	}

	if !h.manager.IsInChannel(req.ChannelID, client) {
		h.sendError(client, "not join this channel")
		return
	}

	message, err := h.messageService.CreateChannelMessage(
		client.UserID,
		req.ChannelID,
		req.Content,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrMessageContentRequired,
		):
			h.sendError(client, err.Error())
		case errors.Is(
			err,
			service.ErrMessageTooLong,
		):
			h.sendError(client, err.Error())
		case errors.Is(
			err,
			service.ErrChannelNotFound,
		):
			h.sendError(client, err.Error())
		case errors.Is(
			err,
			service.ErrNotTeamMember,
		):
			h.sendError(client, err.Error())
		default:
			log.Printf(
				"create channel message failed: %v",
				err,
			)
			h.sendError(client, err.Error())
		}
		return
	}

	outGoingMessage := &ws.OutgoingMessage{
		Type:      "channel_message",
		MessageID: message.ID,
		ChannelID: message.ChannelID,
		UserID:    message.SenderID,
		Content:   message.Content,
		SentAt:    message.CreatedAt.UTC().Format(time.RFC3339),
	}

	result, err := json.Marshal(outGoingMessage)

	if err != nil {
		log.Printf("marshal websocket message failed: %v", err)
		h.sendError(client, "internal server error")
		return
	}

	if err := h.publishChannelMessage(
		req.ChannelID,
		result,
	); err != nil {
		log.Printf(
			"publish channel message failed: user_id=%d channel_id=%d err=%v",
			client.UserID,
			req.ChannelID,
			err,
		)
		return
	}

}

func (h *WSMessageHandler) handleLeaveChannel(
	client *ws.Client,
	req ws.IncomingMessage,
) {
	if req.ChannelID == 0 {
		h.sendError(client, "channel_id is required")
		return
	}

	h.manager.LeaveChannel(req.ChannelID, client)

	h.send(client, ws.OutgoingMessage{
		Type:      "left_channel",
		ChannelID: req.ChannelID,
		Message:   "left channel successfully",
	})

}

func (h *WSMessageHandler) send(
	client *ws.Client,
	message ws.OutgoingMessage,
) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf(
			"marshal websocket message failed: %v\n",
			err,
		)
		return
	}
	client.SendMessage(data)
}

func (h *WSMessageHandler) sendError(
	client *ws.Client,
	message string,

) {
	h.send(client, ws.OutgoingMessage{
		Type:    "error",
		Message: message,
	})
}

func (h *WSMessageHandler) publishChannelMessage(
	channelID uint,
	payload []byte,
) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	return h.broker.PublishChannel(
		ctx,
		channelID,
		payload,
	)
}
