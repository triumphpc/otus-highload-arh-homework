package dao

import (
	"time"

	"otus-highload-arh-homework/internal/social/entity"
)

type Post struct {
	ID        string    `db:"id"`
	AuthorID  int       `db:"author_id"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func FromPostEntity(post entity.Post) Post {
	return Post{
		ID:        post.ID,
		AuthorID:  post.AuthorID,
		Text:      post.Text,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}

func (p Post) ToEntity() entity.Post {
	return entity.Post{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Text:      p.Text,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
