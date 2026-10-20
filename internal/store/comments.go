package store

import (
	"context"
	"database/sql"
)

type CommentAuthor struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
}
type Comment struct {
	Id        int64         `json:"id"`
	PostId    int64         `json:"post_id"`
	UserId    int64         `json:"user_id"`
	Content   string        `json:"content"`
	CreatedAt string        `json:"created_at"`
	User      CommentAuthor `json:"user"`
}

type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {

	query := `
	SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id
	FROM comments c JOIN users u on u.id = c.user_id
	WHERE c.post_id = $1
	ORDER BY c.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, postId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var c Comment
		c.User = CommentAuthor{}
		err = rows.Scan(&c.Id, &c.PostId, &c.UserId, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.Id)

		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil

}
