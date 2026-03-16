package models

import "google.golang.org/protobuf/proto"

type EventType string

const (
	EventNewMessage    = "NEW_MESSAGE"
	EventGetMessage    = "GET_MESSAGE"
	EventListMessages  = "LIST_MESSAGES"
	EventUpdateMessage = "UPDATE_MESSAGE"
	EventDeleteMessage = "DELETE_MESSAGE"
	EventAckMessage    = "ACK_MESSAGE"

	EventGroupCreate       = "GROUP_CREATE"
	EventGroupGet          = "GROUP_GET"
	EventGroupListForUser  = "GROUP_LIST_FOR_USER"
	EventGroupAddMember    = "GROUP_ADD_MEMBER"
	EventGroupRemoveMember = "GROUP_REMOVE_MEMBER"
	EventGroupListMembers  = "GROUP_LIST_MEMBERS"
	EventGroupUpdateRole   = "GROUP_UPDATE_ROLE"
	EventGroupLeave        = "GROUP_LEAVE"
	EventGroupDelete       = "GROUP_DELETE"
)

type EventMessage struct {
	Type      EventType     `json:"type"`
	Payload   proto.Message `json:"payload"`
	Timestamp int64         `json:"timestamp"`
}
