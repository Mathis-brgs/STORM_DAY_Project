package models

import "google.golang.org/protobuf/proto"

type EventType string

const (
	EventNewMessage = "NEW_MESSAGE"
	EventUserTyping = "USER_TYPING"
	EventUserOnline = "USER_ONLINE"
)

type EventMessage struct {
	Type      EventType     `json:"type"`
	Payload   proto.Message `json:"payload"`
	Timestamp int64         `json:"timestamp"`
}
