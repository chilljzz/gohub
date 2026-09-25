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
	manager                    *ws.Manager
	realtimeService            *service.RealtimeService
	messageService             *service.ChannelMessageService
	conversationService        *service.ConversationService
	conversationMessageService *service.ConversationMessageService
	directReceiptService       *service.DirectReceiptService
	broker                     *realtime.RedisBroker
}

func NewWSMessageHandler(
	manager *ws.Manager,
	realtimeService *service.RealtimeService,
	messageService *service.ChannelMessageService,
	conversationService *service.ConversationService,
	conversationMessageService *service.ConversationMessageService,
	directReceiptService *service.DirectReceiptService,
	broker *realtime.RedisBroker,

) *WSMessageHandler {
	return &WSMessageHandler{
		manager:                    manager,
		realtimeService:            realtimeService,
		messageService:             messageService,
		conversationService:        conversationService,
		conversationMessageService: conversationMessageService,
		broker:                     broker,
		directReceiptService:       directReceiptService,
	}
}

func (h *WSMessageHandler) Handle(
	client *ws.Client,
	data []byte,
) {
	var message ws.IncomingMessage

	if err := json.Unmarshal(data, &message); err != nil {
		h.sendError(client, "invalid message format")
		return
	}

	switch message.Type {
	case "join_channel":
		h.handleJoinChannel(client, message)

	case "send_message":
		h.handleSendMessage(client, message)

	case "leave_channel":
		h.handleLeaveChannel(client, message)

	case "join_conversation":
		h.handleJoinConversation(
			client,
			message,
		)

	case "send_conversation_message":
		h.handleSendConversationMessage(
			client,
			message,
		)

	case "leave_conversation":
		h.handleLeaveConversation(
			client,
			message,
		)

	case ws.MessageTypeMarkDelivered:
		h.handleMarkDelivered(
			client,
			message,
		)
	case ws.MessageTypeMarkRead:
		h.handleMarkRead(
			client,
			message,
		)

	default:
		h.sendError(client, "unknown message type")
	}

}

func (h *WSMessageHandler) handleJoinChannel(
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

	conversation, err := h.conversationService.GetChannelConversation(req.ChannelID)

	if err != nil {
		h.handleMessageError(client, err)
		return
	}

	h.manager.JoinConversation(conversation.ID, client)

	userIDs := h.manager.RoomUserIDs(
		conversation.ID,
	)

	log.Printf(
		"conversation room: conversation_id=%d users=%v",
		conversation.ID,
		userIDs,
	)

	h.send(client, ws.OutgoingMessage{
		Type:           "joined_conversation",
		ChannelID:      req.ChannelID,
		ConversationID: conversation.ID,
		Message:        "joined conversation successfully",
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

	conversation, err := h.conversationService.GetChannelConversation(req.ChannelID)

	if err != nil {
		h.handleMessageError(client, err)
		return
	}

	if !h.manager.IsInConversation(conversation.ID, client) {
		h.sendError(client, "not joined conversation")
		return
	}

	messageResult, err := h.messageService.CreateChannelMessage(
		client.UserID,
		req.ChannelID,
		req.Content,
		req.ClientMessageID,
	)

	if err != nil {
		h.handleMessageError(
			client,
			err,
		)

		return
	}

	message := messageResult.Message

	ack := ws.MessageAck{
		Type:            "message_ack",
		ClientMessageID: message.ClientMessageID,
		MessageID:       message.ID,
		ChannelID:       message.ChannelID,
		ConversationID:  conversation.ID,
	}
	ackPayload, err := json.Marshal(ack)
	if err != nil {
		log.Printf(
			"marshal message ack failed: %v",
			err,
		)
		return
	}

	if ok := client.SendMessage(
		ackPayload,
	); !ok {
		log.Printf(
			"send ack failed: user_id=%d message_id=%d",
			client.UserID,
			message.ID,
		)
	}

	if !messageResult.Created {
		return
	}

	outGoingMessage := &ws.OutgoingMessage{
		Type:           "channel_message",
		MessageID:      message.ID,
		ChannelID:      message.ChannelID,
		ConversationID: conversation.ID,
		UserID:         message.SenderID,
		Content:        message.Content,
		SentAt:         message.CreatedAt.UTC().Format(time.RFC3339),
	}

	result, err := json.Marshal(outGoingMessage)

	if err != nil {
		log.Printf("marshal websocket message failed: %v", err)
		h.sendError(client, "internal server error")
		return
	}

	if err := h.publishConversationMessage(
		conversation.ID,
		result,
	); err != nil {
		log.Printf(
			"publish conversation message failed: "+
				"user_id=%d conversation_id=%d err=%v",
			client.UserID,
			conversation.ID,
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

	conversation, err := h.conversationService.GetChannelConversation(req.ChannelID)

	if err != nil {
		h.handleMessageError(client, err)
		return
	}

	h.manager.LeaveConversation(conversation.ID, client)

	h.send(client, ws.OutgoingMessage{
		Type:           "left_conversation",
		ChannelID:      req.ChannelID,
		ConversationID: conversation.ID,
		Message:        "left conversation successfully",
	})

}

func (h *WSMessageHandler) handleJoinConversation(
	client *ws.Client,
	req ws.IncomingMessage,
) {
	if req.ConversationID == 0 {
		h.sendError(
			client,
			"conversation_id is required",
		)
		return
	}

	conversation, err := h.conversationService.
		GetAccessibleConversation(client.UserID, req.ConversationID)
	if err != nil {
		h.handleMessageError(client, err)
		return
	}

	h.manager.JoinConversation(conversation.ID, client)

	h.send(
		client,
		ws.OutgoingMessage{
			Type: "joined_conversation",

			ConversationID: conversation.ID,

			Message: "joined conversation successfully",
		},
	)
}

func (h *WSMessageHandler) handleSendConversationMessage(
	client *ws.Client,
	req ws.IncomingMessage,
) {
	if req.ConversationID == 0 {
		h.sendError(
			client,
			"conversation_id is required",
		)
		return
	}

	if !h.manager.IsInConversation(req.ConversationID, client) {
		h.sendError(
			client,
			"not joined conversation",
		)

		return
	}

	result, err := h.conversationMessageService.CreateMessage(
		client.UserID,
		req.ConversationID,
		req.Content,
		req.ClientMessageID,
	)

	if err != nil {
		h.handleMessageError(
			client,
			err,
		)
		return
	}

	message := result.Message

	ack := ws.MessageAck{
		Type:            "message_ack",
		ClientMessageID: message.ClientMessageID,
		MessageID:       message.ID,
		ConversationID:  message.ConversationID,
	}

	ackPayload, err := json.Marshal(ack)

	if err != nil {
		return
	}

	client.SendMessage(ackPayload)

	if !result.Created {
		return
	}

	outgoing := ws.OutgoingMessage{
		Type:           "conversation_message",
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		UserID:         message.SenderID,
		Content:        message.Content,
		SentAt:         message.CreatedAt.UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(outgoing)

	if err != nil {
		h.sendError(client, "internal server error")
		return
	}

	if err := h.publishConversationMessage(message.ConversationID, payload); err != nil {
		log.Printf(
			"publish conversation message failed: "+
				"user_id=%d conversation_id=%d err=%v",
			client.UserID,
			message.ConversationID,
			err,
		)
	}

}

func (h *WSMessageHandler) handleLeaveConversation(
	client *ws.Client,
	req ws.IncomingMessage,
) {

	if req.ConversationID == 0 {

		h.sendError(
			client,
			"conversation_id is required",
		)

		return
	}

	h.manager.LeaveConversation(
		req.ConversationID,
		client,
	)

	h.send(
		client,
		ws.OutgoingMessage{
			Type: "left_conversation",

			ConversationID: req.ConversationID,

			Message: "left conversation successfully",
		},
	)
}

func (h *WSMessageHandler) handleMarkDelivered(
	client *ws.Client,
	message ws.IncomingMessage,
) {
	result, err := h.directReceiptService.MarkDelivered(
		client.UserID,
		message.ConversationID,
		message.MessageID,
	)

	if err != nil {
		h.sendError(
			client,
			err.Error(),
		)
		return
	}

	h.sendReceiptAck(
		client,
		ws.MessageTypeDelivered,
		result,
	)

	h.publishDirectReceipt(
		ws.MessageTypeDelivered,
		result,
	)

}

func (h *WSMessageHandler) handleMarkRead(
	client *ws.Client,
	message ws.IncomingMessage,
) {
	result, err := h.directReceiptService.MarkRead(
		client.UserID,
		message.ConversationID,
		message.MessageID,
	)

	if err != nil {
		h.sendError(
			client,
			err.Error(),
		)
		return
	}

	h.sendReceiptAck(
		client,
		ws.MessageTypeMarkRead,
		result,
	)

	h.publishDirectReceipt(
		ws.MessageTypeRead,
		result,
	)

}

func (h *WSMessageHandler) sendReceiptAck(
	client *ws.Client,
	receipType string,
	result *service.DirectReceiptResult,
) {
	ack := ws.ReceipAck{
		Type: ws.MessageTypeReceiptAck,

		ReceipType: receipType,

		ConversationID: result.ConversationID,

		MessageID: result.MessageID,
	}

	payload, err := json.Marshal(ack)

	if err != nil {
		return

	}

	client.SendMessage(
		payload,
	)
}

func (h *WSMessageHandler) publishDirectReceipt(
	eventType string,
	result *service.DirectReceiptResult,
) {
	event := ws.ReceiptEvent{
		Type: eventType,

		ConversationID: result.ConversationID,

		MessageID: result.MessageID,

		UserID: result.UserID,

		At: result.At.Format(
			time.RFC3339Nano,
		),
	}

	payload, err := json.Marshal(
		event,
	)
	if err != nil {
		log.Printf(
			"marshal receipt event failed: %v",
			err,
		)
		return
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			time.Second,
		)

	defer cancel()

	if err := h.broker.PublishConversation(
		ctx,
		result.ConversationID,
		payload,
	); err != nil {
		log.Printf(
			"publish receipt failed: conversation_id=%d message_id=%d err=%v",
			result.ConversationID,
			result.MessageID,
			err,
		)

		return
	}

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

func (h *WSMessageHandler) publishConversationMessage(
	conversationID uint,
	payload []byte,
) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	return h.broker.PublishConversation(
		ctx,
		conversationID,
		payload,
	)
}

func (h *WSMessageHandler) handleMessageError(
	client *ws.Client,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrMessageContentRequired),
		errors.Is(err, service.ErrMessageTooLong),
		errors.Is(err, service.ErrChannelNotFound),
		errors.Is(err, service.ErrNotTeamMember),
		errors.Is(err, service.ErrInvalidClientMessageID),
		errors.Is(err, service.ErrClientMessageConflict),
		errors.Is(err, service.ErrConversationNotFound),
		errors.Is(err, service.ErrConversationAccessDenied),
		errors.Is(err, service.ErrInvalidConversationType):

		h.sendError(client, err.Error())

	default:
		log.Printf(
			"websocket message error: %v",
			err,
		)

		h.sendError(
			client,
			"internal server error",
		)
	}
}
