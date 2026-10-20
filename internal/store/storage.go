package store

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
)

type Storage struct {
	Posts interface {
		GetById(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		Delete(context.Context, int64) error
	}
	Users interface {
		Create(context.Context, *User) error
	}

	Comments interface {
		GetByPostId(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostsStore{db},
		Users:    &UsersStore{db},
		Comments: &CommentsStore{db},
	}
}
