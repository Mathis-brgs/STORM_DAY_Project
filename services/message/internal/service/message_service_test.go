package service

import (
	"testing"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/memory"
	"github.com/google/uuid"
)

var (
	testMessageSender   = uuid.MustParse("c1000001-0000-0000-0000-000000000001")
	testMessageReceiver = uuid.MustParse("c1000002-0000-0000-0000-000000000002")
)

func TestMessageServiceMarkMessageReceivedByID(t *testing.T) {
	svc := NewMessageService(memory.NewMessageRepo())

	saved, err := svc.SendMessage(&models.ChatMessage{
		SenderID:       testMessageSender,
		ConversationID: 1,
		Content:        "message to ack",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	acked, err := svc.MarkMessageReceivedByID(saved.ID, testMessageSender, time.Time{})
	if err != nil {
		t.Fatalf("MarkMessageReceivedByID() error = %v", err)
	}
	if acked.ReceivedAt.IsZero() {
		t.Fatalf("expected received_at to be set")
	}

	firstAckUnix := acked.ReceivedAt.Unix()
	overrideTime := time.Unix(1710001234, 0).UTC()
	ackedAgain, err := svc.MarkMessageReceivedByID(saved.ID, testMessageSender, overrideTime)
	if err != nil {
		t.Fatalf("MarkMessageReceivedByID(again) error = %v", err)
	}
	if ackedAgain.ReceivedAt.Unix() != firstAckUnix {
		t.Fatalf("expected first receipt timestamp to be preserved for same user, got %d", ackedAgain.ReceivedAt.Unix())
	}

	receiverTime := time.Unix(1710005678, 0).UTC()
	receiverAck, err := svc.MarkMessageReceivedByID(saved.ID, testMessageReceiver, receiverTime)
	if err != nil {
		t.Fatalf("MarkMessageReceivedByID(receiver) error = %v", err)
	}
	if receiverAck.ReceivedAt.Unix() != receiverTime.Unix() {
		t.Fatalf("expected independent receipt timestamp for second user, got %d", receiverAck.ReceivedAt.Unix())
	}
}

func TestMessageServiceMarkMessageReceivedByID_InvalidID(t *testing.T) {
	svc := NewMessageService(memory.NewMessageRepo())

	if _, err := svc.MarkMessageReceivedByID(0, testMessageSender, time.Now()); err == nil {
		t.Fatalf("expected error for empty id")
	}

	if _, err := svc.MarkMessageReceivedByID(1, uuid.Nil, time.Now()); err == nil {
		t.Fatalf("expected error for empty user id")
	}
}
