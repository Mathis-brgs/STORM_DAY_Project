package subscribers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Mathis-brgs/storm-project/services/notification/internal/service"
	"github.com/nats-io/nats.go"
)

const respondErrorLogFormat = "nats respond error: %v"

type SendRequest struct {
	UserID  string `json:"userId"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type GetRequest struct {
	UserID string `json:"userId"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func StartNotificationSubscribers(nc *nats.Conn, svc *service.NotificationService) error {
	// notification.send — envoyer une notification à un user
	if _, err := nc.Subscribe("notification.send", func(msg *nats.Msg) {
		handleSend(msg, svc)
	}); err != nil {
		return err
	}

	// notification.get — récupérer les notifs non lues d'un user
	if _, err := nc.Subscribe("notification.get", func(msg *nats.Msg) {
		handleGet(msg, svc)
	}); err != nil {
		return err
	}

	// notification.read — marquer toutes les notifs d'un user comme lues
	if _, err := nc.Subscribe("notification.read", func(msg *nats.Msg) {
		handleMarkRead(msg, svc)
	}); err != nil {
		return err
	}

	// Écoute les messages envoyés pour notifier le destinataire
	if _, err := nc.Subscribe("message.sent", func(msg *nats.Msg) {
		handleMessageSent(msg, svc)
	}); err != nil {
		return err
	}

	return nil
}

func handleSend(msg *nats.Msg, svc *service.NotificationService) {
	var req SendRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		respondError(msg, "invalid json")
		return
	}

	notif := service.Notification{
		UserID:  req.UserID,
		Type:    req.Type,
		Payload: req.Payload,
	}

	if err := svc.Send(context.Background(), notif); err != nil {
		respondError(msg, err.Error())
		return
	}

	payload, _ := json.Marshal(map[string]string{"status": "sent"})
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}

func handleGet(msg *nats.Msg, svc *service.NotificationService) {
	var req GetRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		respondError(msg, "invalid json")
		return
	}

	notifs, err := svc.GetPending(context.Background(), req.UserID)
	if err != nil {
		respondError(msg, err.Error())
		return
	}

	payload, _ := json.Marshal(notifs)
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}

func handleMarkRead(msg *nats.Msg, svc *service.NotificationService) {
	var req GetRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		respondError(msg, "invalid json")
		return
	}

	if err := svc.MarkRead(context.Background(), req.UserID); err != nil {
		respondError(msg, err.Error())
		return
	}

	payload, _ := json.Marshal(map[string]string{"status": "ok"})
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}

// handleMessageSent crée automatiquement une notification quand un message est envoyé
type MessageSentEvent struct {
	RecipientID    string `json:"recipientId"`
	SenderUsername string `json:"senderUsername"`
	ConversationID string `json:"conversationId"`
}

func handleMessageSent(msg *nats.Msg, svc *service.NotificationService) {
	var evt MessageSentEvent
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.RecipientID == "" {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"senderUsername": evt.SenderUsername,
		"conversationId": evt.ConversationID,
	})

	notif := service.Notification{
		UserID:  evt.RecipientID,
		Type:    "message",
		Payload: string(payload),
	}

	if err := svc.Send(context.Background(), notif); err != nil {
		log.Printf("notification.send error: %v", err)
	}
}

func respondError(msg *nats.Msg, errMsg string) {
	payload, _ := json.Marshal(ErrorResponse{Error: errMsg})
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}