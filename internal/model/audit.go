package model

import "github.com/google/uuid"

type Action string

const (
	ActionFollow  Action = "follow"
	ActionShorten Action = "shorten"
)

type AuditEvent struct {
	TimeStamp int64      `json:"ts"`
	Action    Action     `json:"action"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	URL       string     `json:"url"`
}
