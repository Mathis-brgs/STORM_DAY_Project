package models

import "google.golang.org/protobuf/proto"

type EventType string

const (
	EventNewMessage    = "NEW_MESSAGE"
	EventGetMessage    = "GET_MESSAGE"
	EventListMessages  = "LIST_MESSAGES"
	EventUpdateMessage = "UPDATE_MESSAGE"
	EventDeleteMessage = "DELETE_MESSAGE"
)

type EventMessage struct {
	Type      EventType     `json:"type"`
	Payload   proto.Message `json:"payload"`
	Timestamp int64         `json:"timestamp"`
}
