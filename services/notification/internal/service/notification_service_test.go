package service

import (
	"context"
	"testing"
)

// ── Send ──────────────────────────────────────────────────────────────────────
// Les tests de validation ne nécessitent pas de connexion Redis réelle.

func TestSend_EmptyUserID(t *testing.T) {
	svc := &NotificationService{rdb: nil}
	err := svc.Send(context.Background(), Notification{
		UserID: "",
		Type:   "message",
	})
	if err == nil {
		t.Fatal("Send with empty userID should return error")
	}
	if err.Error() != "userId requis" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestSend_EmptyType(t *testing.T) {
	svc := &NotificationService{rdb: nil}
	err := svc.Send(context.Background(), Notification{
		UserID: "user-123",
		Type:   "",
	})
	if err == nil {
		t.Fatal("Send with empty type should return error")
	}
	if err.Error() != "type requis" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// ── GetPending ────────────────────────────────────────────────────────────────

func TestGetPending_EmptyUserID(t *testing.T) {
	svc := &NotificationService{rdb: nil}
	_, err := svc.GetPending(context.Background(), "")
	if err == nil {
		t.Fatal("GetPending with empty userID should return error")
	}
	if err.Error() != "userId requis" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// ── MarkRead ──────────────────────────────────────────────────────────────────

func TestMarkRead_EmptyUserID(t *testing.T) {
	svc := &NotificationService{rdb: nil}
	err := svc.MarkRead(context.Background(), "")
	if err == nil {
		t.Fatal("MarkRead with empty userID should return error")
	}
	if err.Error() != "userId requis" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// ── Notification struct ───────────────────────────────────────────────────────

func TestNotification_Defaults(t *testing.T) {
	// Vérifie que la struct se construit correctement
	n := Notification{
		UserID:  "user-456",
		Type:    "message",
		Payload: `{"conversationId":"conv-1"}`,
	}
	if n.Read != false {
		t.Error("Notification.Read should default to false")
	}
	if n.ID != "" {
		t.Error("Notification.ID should be empty before Send")
	}
}

// ── Constants ─────────────────────────────────────────────────────────────────

func TestConstants(t *testing.T) {
	if notifKeyPrefix == "" {
		t.Error("notifKeyPrefix should not be empty")
	}
	if maxPerUser <= 0 {
		t.Error("maxPerUser should be positive")
	}
	if ttl <= 0 {
		t.Error("ttl should be positive")
	}
}
