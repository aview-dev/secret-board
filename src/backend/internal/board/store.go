// 게시글 / 댓글 저장소 (board.db).
//
// 익명성 규칙 (docs/anonymity.md):
//   - 규칙 1: 어떤 테이블에도 작성자 식별자 컬럼을 두지 않는다. 글 내용과 뭉갠 시각만 저장한다.
//   - 규칙 4: 작성 시각은 저장할 때 시간 단위로 내림한다. 정확한 시각은 DB 어디에도 남지 않는다.

package board

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite" // 순수 Go SQLite 드라이버 (CGO 없이 정적 바이너리로 빌드하기 위함)
)

// 글자 수 제한 (바이트가 아니라 글자 기준).
const (
	MaxTitleLen   = 100
	MaxBodyLen    = 10000
	MaxCommentLen = 2000
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
)

type Post struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	CommentCount int       `json:"commentCount"`
}

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"postId"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	db *sql.DB
	// 테스트에서 시각을 고정하려고 바꿔 끼운다.
	now func() time.Time
}

const schema = `
CREATE TABLE IF NOT EXISTS posts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT    NOT NULL,
	body       TEXT    NOT NULL,
	created_at INTEGER NOT NULL -- unix 초, 시간 단위로 내림
);
CREATE TABLE IF NOT EXISTS comments (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id    INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	body       TEXT    NOT NULL,
	created_at INTEGER NOT NULL -- unix 초, 시간 단위로 내림
);
CREATE INDEX IF NOT EXISTS comments_post_id ON comments(post_id, id);
`

// Open 은 path 의 SQLite 파일을 열고 (없으면 만들고) 스키마를 맞춘다.
func Open(path string) (*Store, error) {
	// foreign_keys 는 연결마다 켜야 해서 DSN 에 둔다. busy_timeout 은 쓰기가 겹칠 때 바로 실패하지 않게 한다.
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite 는 쓰기가 한 번에 하나뿐이라 연결 하나로 충분하고, 잠금 충돌도 피한다.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db, now: time.Now}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// blurredNow 는 현재 시각을 시간 단위로 내린 unix 초다 (규칙 4).
func (s *Store) blurredNow() int64 {
	return s.now().UTC().Truncate(time.Hour).Unix()
}

// clean 은 앞뒤 공백을 지우고 1 글자 이상, max 글자 이하인지 확인한다.
func clean(v string, max int) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" || !utf8.ValidString(v) || utf8.RuneCountInString(v) > max {
		return "", ErrInvalid
	}
	return v, nil
}

func (s *Store) CreatePost(ctx context.Context, title, body string) (Post, error) {
	title, err := clean(title, MaxTitleLen)
	if err != nil {
		return Post{}, err
	}
	body, err = clean(body, MaxBodyLen)
	if err != nil {
		return Post{}, err
	}
	at := s.blurredNow()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO posts (title, body, created_at) VALUES (?, ?, ?)`, title, body, at)
	if err != nil {
		return Post{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Post{}, err
	}
	return Post{ID: id, Title: title, Body: body, CreatedAt: time.Unix(at, 0).UTC()}, nil
}

// ListPosts 는 최신 글부터 limit 개를 돌려준다. 본문은 빼고 댓글 수를 붙인다.
func (s *Store) ListPosts(ctx context.Context, limit int) ([]Post, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.created_at, (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id)
		FROM posts p ORDER BY p.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var p Post
		var at int64
		if err := rows.Scan(&p.ID, &p.Title, &at, &p.CommentCount); err != nil {
			return nil, err
		}
		p.CreatedAt = time.Unix(at, 0).UTC()
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *Store) GetPost(ctx context.Context, id int64) (Post, error) {
	var p Post
	var at int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, body, created_at, (SELECT COUNT(*) FROM comments WHERE post_id = posts.id)
		FROM posts WHERE id = ?`, id).Scan(&p.ID, &p.Title, &p.Body, &at, &p.CommentCount)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, err
	}
	p.CreatedAt = time.Unix(at, 0).UTC()
	return p, nil
}

// ListComments 는 오래된 댓글부터 돌려준다.
func (s *Store) ListComments(ctx context.Context, postID int64) ([]Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, post_id, body, created_at FROM comments WHERE post_id = ? ORDER BY id`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		var at int64
		if err := rows.Scan(&c.ID, &c.PostID, &c.Body, &at); err != nil {
			return nil, err
		}
		c.CreatedAt = time.Unix(at, 0).UTC()
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *Store) CreateComment(ctx context.Context, postID int64, body string) (Comment, error) {
	body, err := clean(body, MaxCommentLen)
	if err != nil {
		return Comment{}, err
	}
	at := s.blurredNow()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (post_id, body, created_at)
		 SELECT id, ?, ? FROM posts WHERE id = ?`, body, at, postID)
	if err != nil {
		return Comment{}, err
	}
	// 글이 없으면 SELECT 가 비어서 아무것도 들어가지 않는다.
	if n, _ := res.RowsAffected(); n == 0 {
		return Comment{}, ErrNotFound
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Comment{}, err
	}
	return Comment{ID: id, PostID: postID, Body: body, CreatedAt: time.Unix(at, 0).UTC()}, nil
}
