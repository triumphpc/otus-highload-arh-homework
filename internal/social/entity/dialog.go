package entity

import "time"

// DialogMessage представляет сообщение в диалоге между пользователями
type DialogMessage struct {
	ID         string    `json:"id" db:"id"`
	SenderID   string    `json:"sender_id" db:"sender_id"`
	ReceiverID string    `json:"receiver_id" db:"receiver_id"`
	Text       string    `json:"text" db:"text"`
	SentAt     time.Time `json:"sent_at" db:"sent_at"`
	IsRead     bool      `json:"is_read" db:"is_read"`
}

// Dialog представляет диалог между двумя пользователями
type Dialog struct {
	Participant1 string          `json:"participant1"`
	Participant2 string          `json:"participant2"`
	Messages     []DialogMessage `json:"messages"`
	UnreadCount  int             `json:"unread_count"`
}
