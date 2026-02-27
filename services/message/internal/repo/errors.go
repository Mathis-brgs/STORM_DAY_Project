package repo

import "errors"

var (
	ErrConversationNotFound    = errors.New("conversation not found")
	ErrMembershipNotFound      = errors.New("membership not found")
	ErrMembershipAlreadyExists = errors.New("membership already exists")
)
