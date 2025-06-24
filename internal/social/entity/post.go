package entity

import "time"

type Post struct {
	ID        string    `json:"id"`
	AuthorID  int       `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
