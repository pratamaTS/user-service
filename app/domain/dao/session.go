package dao

import "time"

type Session struct {
	SessionID string    `bson:"session_id" json:"session_id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
	Revoked   bool      `bson:"revoked" json:"revoked"`
	UserAgent string    `bson:"user_agent" json:"user_agent"`
	IP        string    `bson:"ip" json:"ip"`
}
